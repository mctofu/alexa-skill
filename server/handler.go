package server

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mctofu/alexa-skill/alexa"
)

// NewAppHandler returns a http.Handler that services requests using the provided AppHandler
func NewAppHandler(app alexa.AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request", http.StatusBadRequest)
			return
		}

		alexaReq := alexa.RequestEnvelope{}
		err = json.Unmarshal(body, &alexaReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse request: %v", err), http.StatusBadRequest)
			return
		}

		resp, err := app.Handle(r.Context(), &alexaReq)
		if err != nil {
			if _, ok := err.(alexa.ValidationError); ok {
				http.Error(w, fmt.Sprintf("Failed to validate request: %v", err), http.StatusBadRequest)
				return
			}
			http.Error(w, fmt.Sprintf("Unexpected handling error: %v", err), http.StatusInternalServerError)
			return
		}

		out, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to marshal response: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write(out)
	}
}

// CertReader returns the contents of the cert at the provided url
type CertReader func(ctx context.Context, url string) ([]byte, error)

// NewValidatingHandler wraps a http.Handler and validates the request is an authentic request from the
// alexa skill service
func NewValidatingHandler(alexaHandler http.Handler, certReader CertReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cert, err := readValidateCertificate(r, certReader, time.Now())
		if err != nil {
			http.Error(w, fmt.Sprintf("Certificate validation failed: %v", err), http.StatusUnauthorized)
		}

		body, err := readValidateBody(r, cert)
		if err != nil {
			http.Error(w, fmt.Sprintf("Signature validation failed: %v", err), http.StatusUnauthorized)
		}

		// restore request body
		r.Body = ioutil.NopCloser(body)
		alexaHandler.ServeHTTP(w, r)
	}
}

func readValidateCertificate(r *http.Request, certReader CertReader, now time.Time) (*x509.Certificate, error) {
	certURL := r.Header.Get("SignatureCertChainUrl")

	err := verifyCertURL(certURL)
	if err != nil {
		return nil, err
	}

	certContents, err := certReader(r.Context(), certURL)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert at %s: %v", certURL, err)
	}

	block, _ := pem.Decode(certContents)
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	if now.Unix() < cert.NotBefore.Unix() || now.Unix() > cert.NotAfter.Unix() {
		return nil, errors.New("certificate expired")
	}

	foundName := false
	for _, altName := range cert.Subject.Names {
		if altName.Value == "echo-api.amazon.com" {
			foundName = true
		}
	}

	if !foundName {
		return nil, errors.New("certification invalid")
	}

	return cert, nil
}

func readValidateBody(r *http.Request, cert *x509.Certificate) (io.Reader, error) {
	publicKey := cert.PublicKey
	encryptedSig, _ := base64.StdEncoding.DecodeString(r.Header.Get("Signature"))

	var bodyBuf bytes.Buffer
	hash := sha1.New()
	_, err := io.Copy(hash, io.TeeReader(r.Body, &bodyBuf))
	if err != nil {
		return nil, err
	}

	err = rsa.VerifyPKCS1v15(publicKey.(*rsa.PublicKey), crypto.SHA1, hash.Sum(nil), encryptedSig)
	if err != nil {
		return nil, errors.New("signature match failed")
	}

	return &bodyBuf, nil
}

// HTTPCertReader is a CertReader implementation that fetches the contents of the cert
// with a http call. No caching is performed.
func HTTPCertReader(ctx context.Context, certURL string) ([]byte, error) {
	certReq, err := http.NewRequest(http.MethodGet, certURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert request: %v", err)
	}

	certReq = certReq.WithContext(ctx)

	cert, err := http.DefaultClient.Do(certReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get cert: %v", err)
	}
	defer cert.Body.Close()
	certContents, err := ioutil.ReadAll(cert.Body)
	if err != nil {
		return nil, errors.New("could not read Amazon cert file")
	}

	return certContents, nil
}

func verifyCertURL(path string) error {
	link, _ := url.Parse(path)

	if link.Scheme != "https" {
		return errors.New("cert url not https")
	}

	if link.Host != "s3.amazonaws.com" && link.Host != "s3.amazonaws.com:443" {
		return errors.New("cert not on s3")
	}

	if !strings.HasPrefix(link.Path, "/echo.api/") {
		return errors.New("not an echo cert")
	}

	return nil
}
