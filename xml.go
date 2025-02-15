package main

import (
	"encoding/xml"
	"fmt"
	"time"
)

type Feed struct {
	XMLName xml.Name       `xml:"http://www.w3.org/2005/Atom feed"`
	Id      *URI           `xml:"id"`
	Updated *DateConstruct `xml:"updated"`
	Authors []Person       `xml:"author"`
	Title   *TextConstruct `xml:"title"`
	Links   []Link         `xml:"link"`
	Entries []*Entry       `xml:"entry"`

	Categories   []Category `xml:"category"`
	Contributors []Person   `xml:"contributor"`

	Icon      *URI           `xml:"icon"` // 1:1
	Logo      *URI           `xml:"logo"` // 2:1
	Generator *Generator     `xml:"generator"`
	Subtitle  *TextConstruct `xml:"subtitle"`
	Rights    *TextConstruct `xml:"rights"`

	// APP
	Collection *Collection `xml:"http://www.w3.org/2007/app collection"`
}

type Entry struct {
	XMLName xml.Name      `xml:"http://www.w3.org/2005/Atom entry"`
	Id      URI           `xml:"id"`
	Updated DateConstruct `xml:"updated"`
	Authors []Person      `xml:"author"`
	Title   TextConstruct `xml:"title"`
	Links   []Link        `xml:"link"`
	Content Content       `xml:"content"`

	Categories   []Category `xml:"category"`
	Contributors []Person   `xml:"contributor"`

	Published *DateConstruct `xml:"published"`
	Summary   *TextConstruct `xml:"summary"`
	Rights    *TextConstruct `xml:"rights"`
	Source    *Source        `xml:"source"`

	// APP
	Edited  *DateConstruct     `xml:"http://www.w3.org/2007/app edited"`
	Control *PublishingControl `xml:"http://www.w3.org/2007/app control"`
}

// similar to Feed
type Source struct {
	Id      *URI           `xml:"id"`
	Updated *DateConstruct `xml:"updated"`
	Authors []Person       `xml:"author"`
	Title   *TextConstruct `xml:"title"`
	Links   []Link         `xml:"link"`

	Categories   []Category `xml:"category"`
	Contributors []Person   `xml:"contributor"`

	Icon      *URI           `xml:"icon"` // 1:1
	Logo      *URI           `xml:"logo"` // 2:1
	Generator *Generator     `xml:"generator"`
	Subtitle  *TextConstruct `xml:"subtitle"`
	Rights    *TextConstruct `xml:"rights"`

	// APP
	Collection *Collection `xml:"http://www.w3.org/2007/app collection"`
}

type Content struct {
	Type string `xml:"type,attr"` // text, html, xhtml, or mime media type
	Src  string `xml:"src,attr"`
	Body []byte `xml:",innerxml"`
}

type TextConstruct struct {
	XMLName xml.Name
	Type    string `xml:"type,attr"`
	Text    string `xml:",innerxml"`
}

type URI struct {
	XMLName xml.Name
	Target  string `xml:",chardata"`
}

type Link struct {
	Href     string `xml:"href,attr"`
	Relation string `xml:"rel,attr"`  // alternate, related, self, enclosure, via
	Type     string `xml:"type,attr"` // mime media type
	HrefLang string `xml:"hreflang,attr"`
	Title    string `xml:"title,attr"`
	Length   uint64 `xml:"length,attr"`
}

type Category struct {
	Term   string `xml:"term,attr"`
	Scheme string `xml:"scheme,attr"`
	Label  string `xml:"label,attr"`
}

type Generator struct {
	URI     string `xml:"uri,attr"`
	Version string `xml:"version,attr"`
	Text    string `xml:",chardata"`
}

type Person struct {
	XMLName xml.Name
	Name    string `xml:"name"`
	URI     *URI   `xml:"uri"`
}

type DateConstruct struct {
	XMLName xml.Name
	T       time.Time
}

func (t *DateConstruct) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	t.XMLName = start.Name
	if tok, e := d.Token(); e != nil {
		return e
	} else if cd, ok := tok.(xml.CharData); !ok {
		return fmt.Errorf("invalid Atom Date construct")
	} else if updated, e := time.ParseInLocation(time.RFC3339Nano, fmt.Sprintf("%s", cd), nil); e != nil {
		return e
	} else {
		t.T = updated
		return d.Skip()
	}
}

// atom publishing protocol
type Service struct {
	XMLName    xml.Name    `xml:"http://www.w3.org/2007/app service"`
	Workspaces []Workspace `xml:"workspace"`
}

type Workspace struct {
	XMLName     xml.Name      `xml:"http://www.w3.org/2007/app workspace"`
	Title       TextConstruct `xml:"http://www.w3.org/2005/Atom title"`
	Collections []*Collection `xml:"collection"`
}

// The app:collection element MAY appear as a child of an atom:feed or atom:source element in an Atom Feed Document.
// OR may appear in a app:workspace element
type Collection struct {
	XMLName    xml.Name       `xml:"http://www.w3.org/2007/app collection"`
	Href       string         `xml:"href,attr"` // GET to this URI should return an atom:feed document
	Title      *TextConstruct `xml:"http://www.w3.org/2005/Atom title"`
	Categories []Categories   `xml:"categories"`
	Accepts    []Accept       `xml:"accept"`
}

// The content of an "app:accept" element value is a media range as
// defined in [RFC2616].  The media range specifies a type of
// representation that can be POSTed to a Collection.
type Accept struct {
	Text string `xml:",chardata"`
}

type Categories struct {
	XMLName    xml.Name   `xml:"http://www.w3.org/2007/app categories"`
	Fixed      string     `xml:"fixed,attr"`  // yes or no; default no
	Scheme     string     `xml:"scheme,attr"` // URI
	Href       string     `xml:"href,attr"`   // URI pointing to app:categories document
	Categories []Category `xml:"http://www.w3.org/2005/Atom category"`
}

// PublishingControl
type PublishingControl struct {
	XMLName xml.Name `xml:"http://www.w3.org/2007/app control"`
	Draft   string   `xml:"draft"` // yes or no
}
