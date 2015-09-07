// Package raw provides a raw.Data type that is useful
// for writing and reading raw data blobs to and from
// HTTP clients and persistent storage.
package raw

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"sp.com.au/exp/errs"
	"sp.com.au/exp/log"
)

// Maximum size we are prepared to read from a HTTP client.
// Anything this size or larger gets discarded.
var MaxLen = 1024 * 1024 * 16

// Content encodings
const (
	ceIdentity = "identity"
	ceDeflate  = "deflate"
)

// Represents a data BLOB that can be read from or written to
// persistent storage, or a HTTP client.
type Data struct {
	ContentType        string
	ContentEncoding    string
	Content            []byte
	UncompressedLength int
}

// IsCompressed returns whether the content is compressed.
func (data *Data) IsCompressed() bool {
	if data.ContentEncoding == "" {
		data.ContentEncoding = ceIdentity
	}
	return data.ContentEncoding != ceIdentity
}

// ReadRequest reads the data from the request into the raw.Data.
func (data *Data) ReadRequest(ctx context.Context, r *http.Request) error {
	if cl := r.Header.Get("Content-Length"); cl != "" {
		v, err := strconv.ParseInt(cl, 10, 64)
		if err != nil || v < 0 {
			return log.Warn("invalid content-length",
				log.WithValue("content-length", cl),
				log.WithContext(ctx),
				log.WithBadRequest())
		}

		if v >= int64(MaxLen) {
			return log.Warn("max length excceeded",
				log.WithContext(ctx),
				log.WithBadRequest(),
				log.WithValue("MaxLen", MaxLen))
		}

		buf := make([]byte, v)

		_, err = io.ReadFull(r.Body, buf)
		if err != nil {
			return log.Warn("cannot read content",
				log.WithContext(ctx),
				log.WithError(err),
				log.WithBadRequest())
		}
		data.Content = buf
	} else {
		reader := io.LimitReader(r.Body, int64(MaxLen))
		content, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		if len(content) >= MaxLen {
			return log.Warn("max size exceeded",
				log.WithContext(ctx),
				log.WithBadRequest(),
				log.WithValue("MaxLen", MaxLen))
		}
		data.Content = content
	}

	// The HTTP specification does not mention Content-Encoding for
	// requests, but sometimes it is handy to allow the client to do
	// so.
	if ce := r.Header.Get("Content-Encoding"); ce != "" {
		data.ContentEncoding = ce
		data.UncompressedLength = 0 // not known
	} else {
		data.UncompressedLength = len(data.Content)
		data.ContentEncoding = ceIdentity
	}

	data.ContentType = r.Header.Get("Content-Type")
	if data.ContentType == "" {
		data.ContentType = "application/octet-stream"
	}
	return nil
}

// ReadRequest reads the data from the request into the raw.Data.
func (data *Data) ReadResponse(ctx context.Context, r *http.Response) error {
	if cl := r.Header.Get("Content-Length"); cl != "" {
		v, err := strconv.ParseInt(cl, 10, 64)
		if err != nil || v < 0 {
			return log.Warn("invalid content-length",
				log.WithValue("content-length", cl),
				log.WithContext(ctx))
		}

		if v >= int64(MaxLen) {
			return log.Warn("max length excceeded",
				log.WithContext(ctx),
				log.WithBadRequest(),
				log.WithValue("MaxLen", MaxLen))
		}

		buf := make([]byte, v)

		_, err = io.ReadFull(r.Body, buf)
		if err != nil {
			return log.Warn("cannot read content",
				log.WithContext(ctx),
				log.WithError(err))
		}
		data.Content = buf
	} else {
		reader := io.LimitReader(r.Body, int64(MaxLen))
		content, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		if len(content) >= MaxLen {
			return log.Warn("max size exceeded",
				log.WithContext(ctx),
				log.WithBadRequest(),
				log.WithValue("MaxLen", MaxLen))
		}
		data.Content = content
	}

	// The HTTP specification does not mention Content-Encoding for
	// requests, but sometimes it is handy to allow the client to do
	// so.
	if ce := r.Header.Get("Content-Encoding"); ce != "" {
		data.ContentEncoding = ce
		data.UncompressedLength = 0 // not known
	} else {
		data.UncompressedLength = len(data.Content)
		data.ContentEncoding = ceIdentity
	}

	data.ContentType = r.Header.Get("Content-Type")
	if data.ContentType == "" {
		data.ContentType = "application/octet-stream"
	}
	return nil
}

// WriteResponse writes the contents to the client as a response.
func (data *Data) WriteResponse(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// TODO: look at "Accept" header.

	// TODO: this is a very naive handling of the Accept-Encoding
	// header. In particular it does not handle deflate;q=0, which is
	// a valid way of saying that deflate is not acceptable.
	if data.IsCompressed() {
		if ae := r.Header.Get("Accept-Encoding"); !strings.Contains(ae, data.ContentEncoding) {
			// the user agent does not accept the content encoding, so we
			// have to decompress before sending
			err := data.Decompress()
			if err != nil {
				// TODO: not returning to user, should log message and send
				// error to client
				panic("cannot decompress")
			}
		}
	}

	if len(data.Content) == 0 {
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.Header().Del("Content-Encoding")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if data.IsCompressed() {
		w.Header().Set("Content-Encoding", data.ContentEncoding)
	}
	w.Header().Set("Content-Type", data.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data.Content)))
	n, err := w.Write(data.Content)
	if err != nil {
		log.Err(err, log.WithSeverity(log.SeverityWarning))
	}
	if n != len(data.Content) {
		log.Warn("not all bytes sent",
			log.WithValue("expected", len(data.Content)),
			log.WithValue("actual", n))
	}
}

func (data *Data) Decompress() error {
	if !data.IsCompressed() {
		return nil
	}
	if data.ContentEncoding != ceDeflate {
		return log.Error("unknown content-encoding",
			log.WithValue("content-encoding", data.ContentEncoding))
	}
	input := bytes.NewBuffer(data.Content)
	reader := flate.NewReader(input)
	writer := bytes.Buffer{}
	_, err := io.Copy(&writer, reader)
	if err != nil {
		return err
	}
	data.Content = writer.Bytes()
	data.ContentEncoding = ""
	data.UncompressedLength = len(data.Content)
	return nil
}

func (data *Data) Compress() error {
	if data.IsCompressed() || len(data.Content) < 32 {
		// already compressed, or not worth compressing
		// because data is nil or too short
		return nil
	}

	buf := bytes.Buffer{}
	w, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return err
	}
	n, err := w.Write(data.Content)
	if err != nil {
		return err
	}
	if n != len(data.Content) {
		return errs.ServerError("cannot compress")
	}
	err = w.Close()
	if err != nil {
		return err
	}
	compressedBytes := buf.Bytes()

	if len(compressedBytes) < len(data.Content) {
		data.UncompressedLength = len(data.Content)
		data.Content = compressedBytes
		data.ContentEncoding = ceDeflate
	}

	return nil
}

func (data *Data) UnmarshalTo(v interface{}) error {
	err := data.Decompress()
	if err != nil {
		return err
	}
	err = json.Unmarshal(data.Content, v)
	if err != nil {
		return err
	}
	return nil
}

func (data *Data) MarshalFrom(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data.Content = b
	data.ContentType = "application/json"
	data.ContentEncoding = ""
	data.UncompressedLength = len(b)
	return nil
}
