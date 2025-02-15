package main

import (
	"bufio"
	"encoding/base32"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (u *URI) Consumes(v *URI) bool {
	if u == nil && v == nil {
		return true
	} else if u == nil && v != nil || u != nil && v == nil {
		return false
	} else {
		if u.Target == v.Target && u.XMLName == v.XMLName {
			return true
		} else {
			return false
		}
	}
}

func (entry *Entry) UpdatedTime() (t time.Time, err error) {
	if entry == nil {
		err = fmt.Errorf("nil pointer dereference")
	} else {
		t = entry.Updated.T
	}
	return
}

func (d *DateConstruct) Set(t time.Time) (err error) {
	if d == nil {
		err = fmt.Errorf("nil pointer dereference")
	} else {
		d.T = t
	}
	return
}

func (feed *Feed) ETag() (etag string, err error) {
	hasher := fnv.New64a()
	bw := bufio.NewWriter(hasher)
	if feed == nil {
		err = fmt.Errorf("nil pointer dereference")
		return
	} else if e := feed.Id.MarshalTo(bw); e != nil {
		err = e
		return
	} else if e := feed.Updated.MarshalTo(bw); e != nil {
		err = e
		return
	} else if e := bw.Flush(); e != nil {
		err = e
		return
	}
	return base32.StdEncoding.EncodeToString(hasher.Sum(nil))[:13], nil
}

func (entry *Entry) ETag() (etag string, err error) {
	hasher := fnv.New64a()
	bw := bufio.NewWriter(hasher)
	if entry == nil {
		err = fmt.Errorf("nil pointer dereference")
		return
	} else if e := entry.Id.MarshalTo(bw); e != nil {
		err = e
		return
	} else if e := entry.Updated.MarshalTo(bw); e != nil {
		err = e
		return
	} else if e := bw.Flush(); e != nil {
		err = e
		return
	}
	return base32.StdEncoding.EncodeToString(hasher.Sum(nil))[:13], nil
}

// etag
func matchETag(etag string, header_val string) (isSet bool, match bool, err error) {
	if header_val != "" && etag == "" {
		return true, false, nil
	} else if header_val == "" {
		return false, false, nil
	} else if header_val == "*" {
		return true, true, nil
	}
	for _, quote_et := range strings.Split(header_val, ", ") {
		if unquote_et, e := strconv.Unquote(quote_et); e != nil {
			err = e
			return
		} else if unquote_et == etag {
			return true, true, nil
		}
	}
	return true, false, nil
}

func IfMatchIfNoneMatch(etag string, ifmatch string, ifnonematch string) (proceed bool, err error) {
	// If-None-Match
	if isSet, match, e := matchETag(etag, ifnonematch); e != nil {
		err = &HTTPError{
			code:    http.StatusBadRequest,
			message: e.Error(),
		}
		return
	} else if !isSet {
		// continue
	} else if match {
		return
	}
	// If-Match
	if isSet, match, e := matchETag(etag, ifmatch); e != nil {
		err = &HTTPError{
			code:    http.StatusBadRequest,
			message: e.Error(),
		}
		return
	} else if !isSet {
		// continue
	} else if !match {
		return
	}
	proceed = true
	return
}

