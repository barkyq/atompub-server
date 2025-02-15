package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"

	"strconv"

	"testing"
	"time"
)

func TestBackend1(t *testing.T) {
	var tmpdir = "/tmp/gitdir-test"
	h := &Handler{
		B:     NewBackend(NewBillyStorer(tmpdir)),
		gzw:   gzip.NewWriter(nil),
		mutex: new(sync.Mutex),
		buf:   bytes.NewBuffer(nil),
		bw:    bufio.NewWriter(nil),
	}

	defer func() {
		if e := os.RemoveAll(tmpdir); e != nil {
			t.Fatal(e)
		}
	}()

	var feed_to_post_to_root = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<id/>
<title type="text">test microblog</title>
<updated>2025-02-14T10:33:12.546909+01:00</updated>
<author>
<name>John Doe</name>
<uri>mailto:johndoe@example.org</uri>
</author>
<link href="https://example.org/feed.atom" rel="self" type="application/atom+xml"/>
</feed>`
	feed_Path, feed_ETag := func() (string, string) {
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(feed_to_post_to_root))
		req.Header.Set("Content-Type", "application/atom+xml;type=feed")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}
		feed := &Feed{}
		bw := bufio.NewWriter(io.Discard)
		if e := xml.NewDecoder(res.Body).Decode(feed); e != nil {
			t.Fatal(e)
		} else if e := feed.MarshalTo(bw); e != nil {
			t.Fatal(e)
		}

		u, e := url.Parse(res.Header.Get("Location"))
		if e != nil {
			t.Fatal(e)
		}
		etag_quoted := res.Header.Get("ETag")
		etag, e := strconv.Unquote(etag_quoted)
		if e != nil {
			t.Fatal(e)
		}

		return u.Path, etag
	}()

	func() {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		servicedocument := &Service{}
		if e := xml.NewDecoder(res.Body).Decode(servicedocument); e != nil {
			t.Fatal(e)
		} else {
			if servicedocument.XMLName.Local != "service" || servicedocument.XMLName.Space != app_xmlns {
				t.Fatalf("unexpected <service> name")
			} else if ws := servicedocument.Workspaces[0]; ws.XMLName.Local != "workspace" || ws.XMLName.Space != app_xmlns {
				t.Fatalf("unexpected <workspace> name")
			} else if ws.Title.XMLName.Local != "title" || ws.Title.XMLName.Space != atom_xmlns {
				t.Fatalf("unexpected <atom:title> name")
			} else {
				for _, col := range ws.Collections {
					if col == nil {
						continue
					}
					if col.Title == nil {
						t.Fatalf("unset title")
					}
					if u, e := url.Parse(col.Href); e != nil {
						t.Fatal(e)
					} else if u.Path == feed_Path {
						return
					}
				}
				t.Fatal("did not find the collection made previously")
			}
		}
	}()

	var plaintext_to_post_to_feed = `In the realm of code, so sleek and bright,
Lies the slug header, out of sight. mailto:johndoe@example.org
A guide, A:B a map, in URLs it's cast,
Making navigation, oh, so vast. https://example.org/

*Simple*, https://example.org/ yet crucial, in its design,
It helps the web's structure align. sms:15555555555
A title's echo, plain and true,
The slug header, a digital glue.`

	var new_author_name = "Changed Author"
	var new_title = "New Title"
	func() {
		// the feed ETag is based on time.Microsecond
		<-time.After(2 * time.Microsecond)
		req := httptest.NewRequest("POST", feed_Path, bytes.NewBufferString(plaintext_to_post_to_feed))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Slug", "slug header - a digital glue")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}
		entry := &Entry{}
		buf := bytes.NewBuffer(nil)
		bw := bufio.NewWriter(buf)
		if e := xml.NewDecoder(res.Body).Decode(entry); e != nil {
			t.Fatal(e)
		} else if entry.Title.Text = new_title; false {
			//
		} else if entry.Authors = []Person{{
			XMLName: xml.Name{Space: atom_xmlns, Local: "author"},
			Name:    new_author_name,
			URI: &URI{
				XMLName: xml.Name{Space: atom_xmlns, Local: "uri"},
				Target:  "mailto:changed-author@example.org",
			},
		}}; false {
			//
		} else if e := entry.MarshalTo(bw, nil); e != nil {
			t.Fatal(e)
		}

		u, e := url.Parse(res.Header.Get("Location"))
		if e != nil {
			t.Fatal(e)
		} else if e := bw.Flush(); e != nil {
			t.Fatal(e)
		}

		etag_quoted := res.Header.Get("ETag")
		etag, e := strconv.Unquote(etag_quoted)
		if e != nil {
			t.Fatal(e)
		}

		<-time.After(2 * time.Microsecond)
		req = httptest.NewRequest("PUT", u.Path, buf)
		req.Header.Set("Content-Type", "application/atom+xml;type=entry")

		// try with wrong etag
		req.Header.Set("If-Match", strconv.Quote(etag+"X"))
		w = httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res = w.Result()
		if res.StatusCode != http.StatusPreconditionFailed {
			t.Fatal(res.Status)
		}

		// buf should be still OK to read from;
		// the server should NOT read the body since precondition failed
		req = httptest.NewRequest("PUT", u.Path, buf)
		req.Header.Set("Content-Type", "application/atom+xml;type=entry")
		req.Header.Set("If-Match", strconv.Quote(etag))
		w = httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res = w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}
	}()

	// now read the feed and try to find the entry with the updated title and author
	updated_feed_ETag := func() string {
		req := httptest.NewRequest("GET", feed_Path, nil)
		// we changed the feed since we posted to root
		// thus this should return status OK with the feed
		req.Header.Set("If-None-Match", strconv.Quote(feed_ETag))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}
		feed := &Feed{}
		bw := bufio.NewWriter(io.Discard)
		if e := xml.NewDecoder(res.Body).Decode(feed); e != nil {
			t.Fatal(e)
		} else if e := feed.MarshalTo(bw); e != nil {
			t.Fatal(e)
		}
		etag_quoted := res.Header.Get("ETag")
		etag, e := strconv.Unquote(etag_quoted)
		if e != nil {
			t.Fatal(e)
		}

		for _, entry := range feed.Entries {
			if entry.Title.Text == new_title && entry.Authors[0].Name == new_author_name {
				return etag
			}
		}
		t.Fatal("did not find added event!")
		return etag
	}()

	// now change the feed metadata with PUT
	// and get the service document to see if the changes are successfully propagated
	var new_title_2 = "test microblog 2"
	var updated_feed_to_put = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<id/>
<title type="text">test microblog 2</title>
<updated>2025-02-14T10:33:12.546909+01:00</updated>
<author>
<name>John Doe</name>
<uri>mailto:johndoe@example.org</uri>
</author>
<subtitle>This is a Subtitle</subtitle>
<rights>Copyright (C) John Doe 2025</rights>
<icon>https://example.org/favicon.ico</icon>
<logo>https://example.org/logo.png</logo>
<category term="blogposts"/>
<link href="https://example.org/feed.atom" rel="self" type="application/atom+xml"/>
</feed>`
	func() {
		req := httptest.NewRequest("PUT", feed_Path, bytes.NewBufferString(updated_feed_to_put))
		req.Header.Set("Content-Type", "application/atom+xml;type=feed")
		req.Header.Set("If-Match", strconv.Quote(updated_feed_ETag))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}

		req = httptest.NewRequest("GET", "/", nil)
		w = httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res = w.Result()
		servicedocument := &Service{}
		if e := xml.NewDecoder(res.Body).Decode(servicedocument); e != nil {
			t.Fatal(e)
		} else {
			if ws := servicedocument.Workspaces[0]; false {
				//
			} else {
				for _, col := range ws.Collections {
					if col == nil {
						continue
					}
					if u, e := url.Parse(col.Href); e != nil {
						t.Fatal(e)
					} else if u.Path != feed_Path {
						continue
					} else if title := col.Title; title == nil {
						t.Fatalf("unset title")
					} else if title.Text == new_title_2 {
						return
					}
				}
				t.Fatal("did not find the change to the title")
			}
		}
	}()
	if true {
		return
	}
	// now delete everything
	func() {
		req := httptest.NewRequest("DELETE", feed_Path, nil)
		req.Header.Set("If-Match", strconv.Quote(updated_feed_ETag))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}

		req = httptest.NewRequest("GET", "/", nil)
		w = httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res = w.Result()
		servicedocument := &Service{}
		if e := xml.NewDecoder(res.Body).Decode(servicedocument); e != nil {
			t.Fatal(e)
		} else {
			if ws := servicedocument.Workspaces[0]; false {
				//
			} else {
				for _, col := range ws.Collections {
					if col == nil {
						continue
					}
					if u, e := url.Parse(col.Href); e != nil {
						t.Fatal(e)
					} else if u.Path == feed_Path {
						t.Fatalf("expected to be deleted!")
					}
				}
			}
		}
	}()
}

func TestBackend2(t *testing.T) {
	var tmpdir = "/tmp/gitdir-test"
	h := &Handler{
		B:     NewBackend(NewBillyStorer(tmpdir)),
		gzw:   gzip.NewWriter(nil),
		mutex: new(sync.Mutex),
		buf:   bytes.NewBuffer(nil),
		bw:    bufio.NewWriter(nil),
	}

	defer func() {
		if e := os.RemoveAll(tmpdir); e != nil {
			t.Fatal(e)
		}
	}()

	var feed_to_post_to_root = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<id/>
<title type="text">test microblog</title>
<updated>2025-02-14T10:33:12.546909+01:00</updated>
<author>
<name>Jane Doe</name>
<uri>mailto:janedoe@example.org</uri>
</author>
<subtitle>This is a Subtitle</subtitle>
<rights>Copyright (c) Jane Doe 2025</rights>
<icon>https://example.org/favicon.ico</icon>
<category term="blogposts"/>
<link href="https://example.org/feed.atom" rel="self" type="application/atom+xml"/>
</feed>`

	feed_ptr, feed_URL := func() (*Feed, string) {
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(feed_to_post_to_root))
		req.Header.Set("Content-Type", "application/atom+xml;type=feed")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}
		feed := &Feed{}
		if e := xml.NewDecoder(res.Body).Decode(feed); e != nil {
			t.Fatal(e)
		}
		u, e := url.Parse(res.Header.Get("Location"))
		if e != nil {
			t.Fatal(e)
		}
		return feed, u.Path
	}()
	microblogPosts := [3]string{
		"Open protocols are the backbone of a truly decentralized internet, enabling interoperability and freedom from centralized control. mailto:jane@example.org",
		"Embracing open protocols isn't just tech talk; it's about ensuring digital autonomy and innovation for all. #OpenProtocols tel:15555553555",
		"Why do we need open protocols? Because they prevent digital monopolies and foster an ecosystem where everyone can innovate. #Decentralization https://youtube.com/34935js",
	}

	titles := [3]string{
		"The Importance of Open Protocols in Modern Internet Architecture",
		"How Open Protocols Can Liberate Data and Empower Users",
		"The Future of Internet: Why Open Protocols Matter More Than Ever",
	}

	entry_ptrs, entry_URLs := func() (entry_ptrs [3]*Entry, entry_URLs [3]string) {
		for i, slug := range titles {
			req := httptest.NewRequest("POST", feed_URL, bytes.NewBufferString(microblogPosts[i]))
			req.Header.Set("Content-Type", "text/plain")
			req.Header.Set("Slug", slug)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			res := w.Result()
			if res.StatusCode != http.StatusOK {
				t.Fatal(res.Status)
			}
			entry := &Entry{}
			if e := xml.NewDecoder(res.Body).Decode(entry); e != nil {
				t.Fatal(e)
			}
			u, e := url.Parse(res.Header.Get("Location"))
			if e != nil {
				t.Fatal(e)
			}
			entry_ptrs[i] = entry
			entry_URLs[i] = u.Path
		}
		return
	}()

	// simulate restart the handler
	//
	h = &Handler{
		B:     NewBackend(NewBillyStorer(tmpdir)),
		gzw:   gzip.NewWriter(nil),
		mutex: new(sync.Mutex),
		buf:   bytes.NewBuffer(nil),
		bw:    bufio.NewWriter(nil),
	}

	func() {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}
		service := &Service{}
		if e := xml.NewDecoder(res.Body).Decode(service); e != nil {
			t.Fatal(e)
		}

		if len(service.Workspaces[0].Collections) != 1 {
			t.Fatalf("unexpected size of collections")
		}

		col := service.Workspaces[0].Collections[0]
		if col.Title.Text != feed_ptr.Title.Text {
			t.Fatalf("unexpected title of collection")
		} else if col.Href != feed_URL {
			t.Fatalf("unexpected href of collection")
		} else if len(col.Categories) != 1 {
			t.Fatalf("unexpected size of categories")
		} else if cats := col.Categories[0].Categories; len(cats) != 1 {
			t.Fatalf("unexpected size of categories")
		} else if cats[0].Term != "blogposts" {
			t.Fatalf("unexpected category term")
		}
	}()

	func() {
		for i, u := range entry_URLs {
			req := httptest.NewRequest("GET", u, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			res := w.Result()
			if res.StatusCode != http.StatusOK {
				t.Fatal(res.Status)
			}
			entry := &Entry{}
			if e := xml.NewDecoder(res.Body).Decode(entry); e != nil {
				t.Fatal(e)
			}

			old_entry := entry_ptrs[i]

			if entry.Title.Text != old_entry.Title.Text {
				t.Fatalf("unexpected title of entry")
			} else if !entry.Id.Consumes(&old_entry.Id) {
				t.Fatalf("unexpected id of entry")
			} else if !entry.Updated.T.Equal(old_entry.Updated.T) {
				t.Fatalf("unexpected updated time")
			} else if d1, d2 := len(entry.Content.Body), len(old_entry.Content.Body); d1 != d2 {
				t.Fatalf("unexpected content length %d %d", d1, d2)
			} else {
				for j := 0; j < d1; j++ {
					if entry.Content.Body[j] != old_entry.Content.Body[j] {
						t.Fatalf("content does not match at index %d!", j)
					}
				}
			}
		}
	}()

	func() {
		req := httptest.NewRequest("GET", feed_URL, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}
		feed := &Feed{}
		if e := xml.NewDecoder(res.Body).Decode(feed); e != nil {
			t.Fatal(e)
		}

		if feed.Title.Text != feed_ptr.Title.Text {
			t.Fatalf("unexpected title")
		} else if !feed.Id.Consumes(feed_ptr.Id) {
			t.Fatalf("unexpected id of entry")
		} else if feed.Subtitle.Text != feed_ptr.Subtitle.Text {
			t.Fatalf("unexpected subtitle")
		} else if feed.Authors[0].Name != "Jane Doe" {
			t.Fatalf("unexpected author")
		} else if feed.Rights.Text != "Copyright (c) Jane Doe 2025" {
			t.Fatalf("unexpected rights")
		} else if feed.Icon.Target != "https://example.org/favicon.ico" {
			t.Fatalf("unexpected icon")
		} else if len(feed.Links) != 1 {
			t.Fatalf("unexpected links length")
		} else if feed.Links[0].Href != "https://example.org/feed.atom" || feed.Links[0].Relation != "self" {
			t.Fatalf("unexpected self link")
		}
	}()

	func() {
		req := httptest.NewRequest("DELETE", feed_URL, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}
		req = httptest.NewRequest("GET", feed_URL, nil)
		w = httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res = w.Result()
		if res.StatusCode != http.StatusNotFound {
			t.Fatal(res.Status)
		}
	}()

	// simulate restart the handler
	//
	h = &Handler{
		B:     NewBackend(NewBillyStorer(tmpdir)),
		gzw:   gzip.NewWriter(nil),
		mutex: new(sync.Mutex),
		buf:   bytes.NewBuffer(nil),
		bw:    bufio.NewWriter(nil),
	}

	func() {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		res := w.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatal(res.Status)
		}
		service := &Service{}
		if e := xml.NewDecoder(res.Body).Decode(service); e != nil {
			t.Fatal(e)
		}

		if len(service.Workspaces[0].Collections) != 0 {
			t.Fatalf("unexpected size of collections after deletion")
		}

		for _, u := range entry_URLs {
			req := httptest.NewRequest("GET", u, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			res := w.Result()
			if res.StatusCode != http.StatusNotFound {
				t.Fatal(res.Status)
			}
		}
	}()
}
