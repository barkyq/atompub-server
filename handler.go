package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
)

type IBackend interface {
	GetRoot(r *http.Request) (sd *Service, err *HTTPError)
	PostToRoot(r *http.Request, new_feed *Feed) (feed *Feed, feed_URL string, err *HTTPError)

	GetFeed(r *http.Request) (feed *Feed, err *HTTPError)
	PostToFeed(r *http.Request) (entry *Entry, entry_URL string, err *HTTPError)
	PutFeed(r *http.Request, new_feed *Feed) (err *HTTPError)
	DeleteFeed(r *http.Request) (err *HTTPError)

	GetEntry(r *http.Request) (entry *Entry, err *HTTPError)
	PutEntry(r *http.Request, new_entry *Entry) (err *HTTPError)
	DeleteEntry(r *http.Request) (err *HTTPError)

	GetMedia(r *http.Request) (media []byte, mediatype string, err *HTTPError)
}

type Handler struct {
	B     IBackend
	gzw   *gzip.Writer
	mutex *sync.Mutex
	buf   *bytes.Buffer
	bw    *bufio.Writer
}

type HTTPError struct {
	code    int
	message string
}

func (e *HTTPError) Error() string {
	if e.message != "" {
		return http.StatusText(e.code) + " [note: " + e.message + "]"
	} else {
		return http.StatusText(e.code)
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only allow one connection at a time
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var err error
	var body []byte
	switch path.Dir(r.URL.Path) {
	case "/":
		switch r.Method {
		case "OPTIONS":
			w.Header().Add("Allow", "OPTIONS, GET, POST")
			w.WriteHeader(http.StatusOK)
			return
		case "GET":
			body, err = h.serveRoot(w, r)
		case "POST":
			body, err = h.postToRoot(w, r)
		default:
			err = &HTTPError{code: http.StatusMethodNotAllowed}
		}
	case "/feed":
		switch r.Method {
		case "OPTIONS":
			w.Header().Add("Allow", "OPTIONS, GET, POST, PUT, DELETE")
			w.WriteHeader(http.StatusOK)
			return
		case "GET":
			body, err = h.serveFeed(w, r)
		case "POST":
			body, err = h.postToFeed(w, r)
		case "PUT":
			err = h.putFeed(w, r)
		case "DELETE":
			err = h.deleteFeed(w, r)
		default:
			err = &HTTPError{code: http.StatusMethodNotAllowed}
		}
	case "/entry":
		switch r.Method {
		case "OPTIONS":
			w.Header().Add("Allow", "OPTIONS, GET, PUT, DELETE")
			w.WriteHeader(http.StatusOK)
			return
		case "GET":
			body, err = h.serveEntry(w, r)
		case "PUT":
			err = h.putEntry(w, r)
		case "DELETE":
			err = h.deleteEntry(w, r)
		default:
			err = &HTTPError{code: http.StatusMethodNotAllowed}
		}
	case "/media":
		switch r.Method {
		case "OPTIONS":
			w.Header().Add("Allow", "OPTIONS, GET, PUT")
			// do not allow DELETION of /media/X
			// should instead delete /entry/X
			w.WriteHeader(http.StatusOK)
			return
		case "GET":
			body, err = h.serveMedia(w, r)
		case "PUT":
			// edit media
			err = &HTTPError{code: http.StatusNotImplemented}
		default:
			err = &HTTPError{code: http.StatusMethodNotAllowed}
		}
	default:
		err = &HTTPError{code: http.StatusMethodNotAllowed}
	}
	if e, ok := err.(*HTTPError); ok {
		http.Error(w, e.Error(), e.code)
	} else if err != nil {
		// 500 internal server error
		http.Error(w, http.StatusText(500), 500)
	} else if body == nil {
		// nil body but no error means handler already responded
		// with call to WriteHeader
		return
	} else if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		h.gzw.Reset(w)
		h.gzw.Write(body)
		h.gzw.Close()
	} else {
		w.Write(body)
	}
}

// PUT

func (h *Handler) putFeed(w http.ResponseWriter, r *http.Request) (err error) {
	if r.Header.Get("Content-Type") != "application/atom+xml;type=feed" {
		err = &HTTPError{
			code:    http.StatusUnsupportedMediaType,
			message: "content-type must be application/atom+xml;type=feed",
		}
		return
	}

	new_feed := &Feed{}
	if feed, e := h.B.GetFeed(r); e != nil {
		err = e
	} else if etag, e := feed.ETag(); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
	} else if proceed, e := IfMatchIfNoneMatch(etag, r.Header.Get("If-Match"), r.Header.Get("If-None-Match")); e != nil {
		err = e
	} else if !proceed {
		err = &HTTPError{code: http.StatusPreconditionFailed}
	} else if e := xml.NewDecoder(r.Body).Decode(new_feed); e != nil {
		return &HTTPError{code: http.StatusBadRequest, message: "could not unmarshal request body"}
	} else if new_feed.Id = feed.Id; false {
		//
	} else if new_feed.Updated = feed.Updated; false {
		//
	} else if e := new_feed.Validate(); e != nil {
		err = &HTTPError{code: http.StatusBadRequest, message: e.Error()}
	} else if e := h.B.PutFeed(r, new_feed); e != nil {
		err = e
	}

	if err != nil {
		return
	} else {
		w.WriteHeader(http.StatusOK)
		return nil
	}
}

func (h *Handler) putEntry(w http.ResponseWriter, r *http.Request) (err error) {
	if r.Header.Get("Content-Type") != "application/atom+xml;type=entry" {
		err = &HTTPError{code: http.StatusUnsupportedMediaType, message: "content-type must be application/atom+xml;type=entry"}
		return
	}

	new_entry := &Entry{}
	if entry, e := h.B.GetEntry(r); e != nil {
		err = e
	} else if etag, e := entry.ETag(); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
	} else if proceed, e := IfMatchIfNoneMatch(etag, r.Header.Get("If-Match"), r.Header.Get("If-None-Match")); e != nil {
		err = e
	} else if !proceed {
		err = &HTTPError{code: http.StatusPreconditionFailed}
	} else if e := xml.NewDecoder(r.Body).Decode(new_entry); e != nil {
		return &HTTPError{code: http.StatusBadRequest, message: "could not unmarshal request body"}
	} else if new_entry.Source = entry.Source; false {
		// cannot change the source
	} else if _, e := new_entry.Validate(nil); e != nil {
		err = &HTTPError{code: http.StatusBadRequest, message: "invalid atom entry"}
	} else if e := h.B.PutEntry(r, new_entry); e != nil {
		err = e
	}
	if err != nil {
		return
	} else {
		w.WriteHeader(http.StatusOK)
		return nil
	}
}

// POST

// body is <atom:feed> specifying metadata for new collection
// handler is lax about missing "required" fields, such as updated, id, etc
// response is <atom:feed>, presumably without entries, containing an <app:collection> child
func (h *Handler) postToRoot(w http.ResponseWriter, r *http.Request) (body []byte, err error) {
	if r.Header.Get("Content-Type") != "application/atom+xml;type=feed" {
		err = &HTTPError{
			code:    http.StatusUnsupportedMediaType,
			message: "content-type must be application/atom+xml;type=feed",
		}
		return
	}

	new_feed := &Feed{}
	if e := xml.NewDecoder(r.Body).Decode(new_feed); e != nil {
		err = &HTTPError{code: http.StatusBadRequest, message: "could not unmarshal request body"}
	} else if e := new_feed.Validate(); e != nil {
		err = &HTTPError{code: http.StatusBadRequest, message: e.Error()}
	} else if feed, feed_URL, e := h.B.PostToRoot(r, new_feed); e != nil {
		err = e
	} else if etag, e := feed.ETag(); e != nil {
		err = e
	} else if h.buf.Reset(); false {
		//
	} else if _, e := h.buf.WriteString(xml.Header); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		//
	} else if h.bw.Reset(h.buf); false {
		//
	} else if e := feed.MarshalTo(h.bw); e != nil {
		err = e
	} else if e := h.bw.Flush(); e != nil {
		err = e
	} else if w.Header().Set("Content-Type", "application/atom+xml;type=feed"); false {
		//
	} else if w.Header().Set("Location", feed_URL); false {
		//
	} else if w.Header().Set("ETag", strconv.Quote(etag)); false {
		//
	} else {
		body = h.buf.Bytes()
	}
	return
}

func (h *Handler) postToFeed(w http.ResponseWriter, r *http.Request) (body []byte, err error) {
	err = &HTTPError{code: http.StatusInternalServerError}
	if entry, entry_URL, e := h.B.PostToFeed(r); e != nil {
		// could be not found, or something else
		err = e
		return
	} else if etag, e := entry.ETag(); e != nil {
		return
	} else if h.buf.Reset(); false {
		//
	} else if _, e := h.buf.WriteString(xml.Header); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		//
	} else if h.bw.Reset(h.buf); false {
		//
	} else if e := entry.MarshalTo(h.bw, nil); e != nil {
		return
	} else if e := h.bw.Flush(); e != nil {
		return
	} else if w.Header().Set("Location", entry_URL); false {
		//
	} else if w.Header().Set("Content-Type", "application/atom+xml;type=entry"); false {
		//
	} else if w.Header().Set("ETag", strconv.Quote(etag)); false {
		//
	} else {
		return h.buf.Bytes(), nil
	}
	return
}

// GET

func (h *Handler) serveRoot(w http.ResponseWriter, r *http.Request) (body []byte, err error) {
	err = &HTTPError{code: http.StatusInternalServerError}
	if sd, e := h.B.GetRoot(r); e != nil {
		// could be not found, or something else
		err = e
		return
	} else if h.buf.Reset(); false {
		//
	} else if _, e := h.buf.WriteString(xml.Header); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		//
	} else if h.bw.Reset(h.buf); false {
		//
	} else if e := sd.MarshalTo(h.bw); e != nil {
		return
	} else if e := h.bw.Flush(); e != nil {
		return
	} else if w.Header().Set("Content-Type", "application/atomsvc+xml"); false {
		//
	} else {
		return h.buf.Bytes(), nil
	}
	return
}

func (h *Handler) serveFeed(w http.ResponseWriter, r *http.Request) (body []byte, err error) {
	err = &HTTPError{code: http.StatusInternalServerError}
	if feed, e := h.B.GetFeed(r); e != nil {
		// could be not found, or something else
		err = e
		return
	} else if etag, e := feed.ETag(); e != nil {
		return
	} else if proceed, e := IfMatchIfNoneMatch(etag, r.Header.Get("If-Match"), r.Header.Get("If-None-Match")); e != nil {
		// could be bad request
		err = e
		return
	} else if !proceed {
		// should return not modified
		w.WriteHeader(http.StatusNotModified)
		// return empty body and nil error signaling that response already written
		return nil, nil
	} else if h.buf.Reset(); false {
		//
	} else if _, e := h.buf.WriteString(xml.Header); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		//
	} else if h.bw.Reset(h.buf); false {
		//
	} else if e := feed.MarshalTo(h.bw); e != nil {
		return
	} else if e := h.bw.Flush(); e != nil {
		return
	} else if w.Header().Set("Content-Type", "application/atom+xml;type=feed"); false {
		//
	} else if w.Header().Set("ETag", strconv.Quote(etag)); false {
		//
	} else {
		return h.buf.Bytes(), nil
	}
	return
}

func (h *Handler) serveEntry(w http.ResponseWriter, r *http.Request) (body []byte, err error) {
	err = &HTTPError{code: http.StatusInternalServerError}
	if entry, e := h.B.GetEntry(r); e != nil {
		// could be not found, or something else
		err = e
		return
	} else if etag, e := entry.ETag(); e != nil {
		return
	} else if proceed, e := IfMatchIfNoneMatch(etag, r.Header.Get("If-Match"), r.Header.Get("If-None-Match")); e != nil {
		// could be bad request
		err = e
		return
	} else if !proceed {
		w.WriteHeader(http.StatusNotModified)
		return nil, nil
	} else if h.buf.Reset(); false {
		//
	} else if _, e := h.buf.WriteString(xml.Header); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		//
	} else if h.bw.Reset(h.buf); false {
		//
	} else if e := entry.MarshalTo(h.bw, nil); e != nil {
		return
	} else if e := h.bw.Flush(); e != nil {
		return
	} else if w.Header().Set("Content-Type", "application/atom+xml;type=entry"); false {
		//
	} else if w.Header().Set("ETag", strconv.Quote(etag)); false {
		//
	} else {
		return h.buf.Bytes(), nil
	}
	return
}

func (h *Handler) serveMedia(w http.ResponseWriter, r *http.Request) (body []byte, err error) {
	err = &HTTPError{code: http.StatusInternalServerError}
	if media, mediatype, e := h.B.GetMedia(r); e != nil {
		// could be not found, or something else
		err = e
		return
	} else if w.Header().Set("Content-Type", mediatype); false {
		//
	} else {
		return media, nil
	}
	return
}

// DELETE

func (h *Handler) deleteEntry(w http.ResponseWriter, r *http.Request) (err error) {
	err = &HTTPError{code: http.StatusInternalServerError}
	if entry, e := h.B.GetEntry(r); e != nil {
		err = e
	} else if etag, e := entry.ETag(); e != nil {
		//
	} else if proceed, e := IfMatchIfNoneMatch(etag, r.Header.Get("If-Match"), r.Header.Get("If-None-Match")); e != nil {
		err = e
	} else if !proceed {
		err = &HTTPError{code: http.StatusPreconditionFailed}
	} else if e := h.B.DeleteEntry(r); e != nil {
		err = e
	} else {
		w.WriteHeader(http.StatusOK)
		return nil
	}
	return
}

func (h *Handler) deleteFeed(w http.ResponseWriter, r *http.Request) (err error) {
	err = &HTTPError{code: http.StatusInternalServerError}
	if feed, e := h.B.GetFeed(r); e != nil {
		err = e
	} else if etag, e := feed.ETag(); e != nil {
		//
	} else if proceed, e := IfMatchIfNoneMatch(etag, r.Header.Get("If-Match"), r.Header.Get("If-None-Match")); e != nil {
		err = e
	} else if !proceed {
		err = &HTTPError{code: http.StatusPreconditionFailed}
	} else if e := h.B.DeleteFeed(r); e != nil {
		err = e
	} else {
		w.WriteHeader(http.StatusOK)
		return nil
	}
	return
}
