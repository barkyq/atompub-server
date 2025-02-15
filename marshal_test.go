package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"testing"
	"time"
)

func TestMarshalEntry(t *testing.T) {
	author := Person{
		XMLName: xml.Name{Space: atom_xmlns, Local: "author"},
		Name:    "John Doe",
		URI:     &URI{XMLName: xml.Name{Space: atom_xmlns, Local: "uri"}, Target: "https://example.org"},
	}

	now := DateConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "updated"},
		T:       time.Now().Round(time.Second),
	}

	rights := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "rights"},
		Text:    "Copyright (c) 2025, John Doe",
	}

	title := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
		Text:    "February Blog Post",
	}

	link := Link{
		Href:     "https://example.org/atom/feed",
		Relation: "self",
		Type:     "application/atom+xml",
	}

	category := Category{
		Term: "blog posts",
	}

	content := Content{
		Type: "text",
		Body: []byte("Abstract. Given a Hamiltonian isotopy of a symplectic manifold, a spectral invariant (or action selector) is a certain distinguished critical value of the Hamiltonian action functional. In a pioneering 1992 paper, Viterbo introduces a spectral metric for a Hamiltonian isotopy as a certain combination of two spectral invariants. Subsequently this metric was reimplemented by Schwarz and Oh for a wide class of symplectic manifolds using the technology of Floer homology. This raises the question of contact geometric analogues. There are some notable contributions to this subject: Givental's 1990s work and Sandon's 2010s work on generating functions and Albers-Merry's 2018 work on spectral invariants via Rabinowitz-Floer homology. In my talk, I will present the construction of one spectral metric for each Reeb flow and each positive loop (on certain contact manifolds amenable to Floer theory), and will discuss the relation to contact displaceability, orderability, and the existence of closed Reeb orbits.\n\n1. Spectral invariants for Hamiltonian isotopies.\n\n1.1. Key features. Spectrality, continuity, conjugation invariance.\n\n1.2. History. Viterbo 1992, Schwarz 2000, Oh 2005.\n\n1.3. Applications. Non-squeezing.\n\n1.4. Spectral norm. Interesting metric structure, giving lower bound on Hofer geometry. Appears in displacement energy bound. \n\nRemark. If M is aspherical, then depends only on time-1 map because gamma-norm vanishes on loops.\n\nRemark. Work of Humilière, Jannaud, Leclercq on degeneracy of spectral norm on the universal cover.\n\n2. Contact geometry analogue\n\n2.1. What is the correct spectrum for a contact isotopy.\n\nThe literature mainly focuses on the spectrum of lengths of translated points — this depends on an auxiliary Reeb flow.\n\nSandon constructs GF whose critical values are lengths of translated points. RFH constructs an action functional whose critical values are lengths of translated points.\n\nRemark. SFT action functional spectrum are lengths of closed Reeb orbits. This is the spectrum of the identity.\n\nRemark. Weinstein's and Sandon's conjecture.\n\n2.2. The Discriminant. Givental's 1990 work suggests a general perspective using the \"discriminant\"\n\nThis is a codimension 1 set in the space of contact isotopies. \n\nReformulation of Sandon's conjecture and Weinstein's conjecture.\n\n2.3. Discriminant functors. Category C of contact isotopies with morphisms being non-negative paths. Discriminant functor are functors from C to the category of vector spaces satisfying the axiom:\n\n(✨) HF(morphism) is isomorphism if morphism does not cross the discriminant.\n\nRemark. Why non-negativity? Why contact isotopies? This conditions arise in the Floer theory construction.\n\n2.4. Colimit of HF is SH. To non-eternal classes in SH one can associate real-valued spectral invariants. Depend only on the projection to the universal cover.\n\n2.5. Reeb flow variant; these are valued in the aforementioned spectrum. Comment that these are not conjugation invariant. Give brief indication why.\n\n2.6. Positive loop variant. These are integer valued. Comment that they are conjugation invariant and explain why.\n\n2.7. Comment a mild conjugation invariance.\n\n3. Product structures.\n\n3.1. Definition. Natural transformation. Induced product on SH.\n\nExample. The pair of pants product.\n\n3.2. Comment sub-additivity of spectral invariants. Give brief explanation.\n\n4. Non-eternal idempotents and spectral metrics.\n\n4.1. Idempotent in SH which lies in image of HF(id) and is not eternal.\n\nThen c(id) = 0. Sub-additivity gives a pseudometric. If one uses the loop variant, then the resulting metric is conjugation invariant (but integer valued).\n\n4.2. Displacement trick. Metric is non-degenerate on the quotient.\n\n4.3. Growth rate of positive loops and the systole.\n\n4.4. Weinstein conjecture and non-degeneracy of the spectral norm.\n"),
	}

	entry := &Entry{
		Id: URI{
			XMLName: xml.Name{Space: atom_xmlns, Local: "id"},
			Target:  "https://example.org/atom/feed",
		},
		Authors:    []Person{author},
		Title:      title,
		Updated:    now,
		Rights:     &rights,
		Links:      []Link{link},
		Categories: []Category{category},
		Content:    content,
	}

	bw := bufio.NewWriter(io.Discard)
	if e := entry.MarshalTo(bw, nil); e != nil {
		t.Fatal(e)
	} else if e := bw.Flush(); e != nil {
		t.Fatal(e)
	}
}

const blog_post_raw = "Abstract. Given a Hamiltonian isotopy of a symplectic manifold, a spectral invariant (or action selector) is a certain distinguished critical value of the Hamiltonian action functional. In a pioneering 1992 paper, Viterbo introduces a spectral metric for a Hamiltonian isotopy as a certain combination of two spectral invariants. Subsequently this metric was reimplemented by Schwarz and Oh for a wide class of symplectic manifolds using the technology of Floer homology. This raises the question of contact geometric analogues. There are some notable contributions to this subject: Givental's 1990s work and Sandon's 2010s work on generating functions and Albers-Merry's 2018 work on spectral invariants via Rabinowitz-Floer homology. In my talk, I will present the construction of one spectral metric for each Reeb flow and each positive loop (on certain contact manifolds amenable to Floer theory), and will discuss the relation to contact displaceability, orderability, and the existence of closed Reeb orbits.\n\n1. Spectral invariants for Hamiltonian isotopies.\n\n1.1. Key features. Spectrality, continuity, conjugation invariance.\n\n1.2. History. Viterbo 1992, Schwarz 2000, Oh 2005.\n\n1.3. Applications. Non-squeezing.\n\n1.4. Spectral norm. Interesting metric structure, giving lower bound on Hofer geometry. Appears in displacement energy bound. \n\nRemark. If M is aspherical, then depends only on time-1 map because gamma-norm vanishes on loops.\n\nRemark. Work of Humilière, Jannaud, Leclercq on degeneracy of spectral norm on the universal cover.\n\n2. Contact geometry analogue\n\n2.1. What is the correct spectrum for a contact isotopy.\n\nThe literature mainly focuses on the spectrum of lengths of translated points — this depends on an auxiliary Reeb flow.\n\nSandon constructs GF whose critical values are lengths of translated points. RFH constructs an action functional whose critical values are lengths of translated points.\n\nRemark. SFT action functional spectrum are lengths of closed Reeb orbits. This is the spectrum of the identity.\n\nRemark. Weinstein's and Sandon's conjecture.\n\n2.2. The Discriminant. Givental's 1990 work suggests a general perspective using the \"discriminant\"\n\nThis is a codimension 1 set in the space of contact isotopies. \n\nReformulation of Sandon's conjecture and Weinstein's conjecture.\n\n2.3. Discriminant functors. Category C of contact isotopies with morphisms being non-negative paths. Discriminant functor are functors from C to the category of vector spaces satisfying the axiom:\n\n(✨) HF(morphism) is isomorphism if morphism does not cross the discriminant.\n\nRemark. Why non-negativity? Why contact isotopies? This conditions arise in the Floer theory construction.\n\n2.4. Colimit of HF is SH. To non-eternal classes in SH one can associate real-valued spectral invariants. Depend only on the projection to the universal cover.\n\n2.5. Reeb flow variant; these are valued in the aforementioned spectrum. Comment that these are not conjugation invariant. Give brief indication why.\n\n2.6. Positive loop variant. These are integer valued. Comment that they are conjugation invariant and explain why.\n\n2.7. Comment a mild conjugation invariance.\n\n3. Product structures.\n\n3.1. Definition. Natural transformation. Induced product on SH.\n\nExample. The pair of pants product.\n\n3.2. Comment sub-additivity of spectral invariants. Give brief explanation.\n\n4. Non-eternal idempotents and spectral metrics.\n\n4.1. Idempotent in SH which lies in image of HF(id) and is not eternal.\n\nThen c(id) = 0. Sub-additivity gives a pseudometric. If one uses the loop variant, then the resulting metric is conjugation invariant (but integer valued).\n\n4.2. Displacement trick. Metric is non-degenerate on the quotient.\n\n4.3. Growth rate of positive loops and the systole.\n\n4.4. Weinstein conjecture and non-degeneracy of the spectral norm.\n"

func TestMarshalFeed(t *testing.T) {
	author := Person{
		XMLName: xml.Name{Space: atom_xmlns, Local: "author"},
		Name:    "John Doe",
		URI:     &URI{XMLName: xml.Name{Space: atom_xmlns, Local: "uri"}, Target: "https://example.org"},
	}

	now := DateConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "updated"},
		T:       time.Now().Round(time.Second),
	}

	rights := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "rights"},
		Text:    "Copyright (c) 2025, John Doe",
	}

	generator := Generator{
		URI:     "https://example.org/feed-generator",
		Version: "v0.0.1-alpha",
	}

	feed := &Feed{
		Id: &URI{
			XMLName: xml.Name{Space: atom_xmlns, Local: "id"},
			Target:  "https://example.org/atom/feed",
		},
		Icon: &URI{
			XMLName: xml.Name{Space: atom_xmlns, Local: "icon"},
			Target:  "https://example.org/atom/icon.png",
		},
		Logo: &URI{
			XMLName: xml.Name{Space: atom_xmlns, Local: "logo"},
			Target:  "https://example.org/atom/logo.png",
		},
		Authors:   []Person{author},
		Updated:   &now,
		Rights:    &rights,
		Generator: &generator,
	}

	feed.Title = &TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
		Type:    "html",
		Text:    "John Doe Blog",
	}

	link := Link{
		Href:     "https://example.org/atom/feed",
		Relation: "self",
		Type:     "application/atom+xml",
	}

	feed.Links = []Link{link}

	category := Category{
		Term: "blog posts",
	}

	feed.Categories = []Category{category}
	title := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
		Text:    "February Blog Post",
	}
	buf := bytes.NewBuffer(nil)
	xml.EscapeText(buf, []byte(blog_post_raw))
	content := Content{
		Type: "text",
		Body: buf.Bytes(),
	}

	entry := &Entry{
		Id: URI{
			XMLName: xml.Name{Space: atom_xmlns, Local: "id"},
			Target:  "https://example.org/atom/feed",
		},
		Authors:    []Person{author},
		Title:      title,
		Updated:    now,
		Rights:     &rights,
		Links:      []Link{link},
		Categories: []Category{category},
		Content:    content,
	}
	feed.Entries = []*Entry{entry}

	bw := bufio.NewWriter(io.Discard)
	if e := feed.MarshalTo(bw); e != nil {
		t.Fatal(e)
	} else if e := bw.Flush(); e != nil {
		t.Fatal(e)
	}
}

func TestMarshalFeedWithCollection(t *testing.T) {
	author := Person{
		XMLName: xml.Name{Space: atom_xmlns, Local: "author"},
		Name:    "John Doe",
		URI:     &URI{XMLName: xml.Name{Space: atom_xmlns, Local: "uri"}, Target: "mailto:johndoe@example.org"},
	}
	fixed_date := time.Date(2025, time.February, 9, 16, 20, 0, 0, time.Local)
	updated_feed := DateConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "updated"},
		T:       fixed_date,
	}
	rights := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "rights"},
		Text:    "Copyright (c) 2025, John Doe",
	}
	feed_link := Link{
		Href:     "https://example.org/feed",
		Relation: "self",
		Type:     "application/atom+xml",
	}
	category := Category{
		Term: "blog posts",
	}
	feed_title := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
		Type:    "html",
		Text:    "John Doe Blog",
	}
	feed := &Feed{
		Id: &URI{
			XMLName: xml.Name{Space: atom_xmlns, Local: "id"},
			Target:  "https://example.org/feed/",
		},
		Authors:    []Person{author},
		Updated:    &updated_feed,
		Rights:     &rights,
		Links:      []Link{feed_link},
		Categories: []Category{category},
		Title:      &feed_title,
	}

	feed.Entries = make([]*Entry, 3)
	uuids := [3]string{
		"urn:uuid:59592fc2-0a7d-47fb-aceb-7d4a7bd6985a",
		"urn:uuid:59592fc2-0a7d-47fb-aceb-7d4a7bd6985b",
		"urn:uuid:59592fc2-0a7d-47fb-aceb-7d4a7bd6985c",
	}
	titles := [3]string{
		"Random Thoughts 3",
		"Random Thoughts 2",
		"Random Thoughts 1",
	}
	bodies := [3]string{
		"Today's musings on the absurdity of life, the universe, and why my coffee never tastes the same twice. #LifeQuestions #CoffeeConundrums",
		"Thoughts on how we chase perfection but often miss the beauty in imperfection. Embracing flaws might just be the key to happiness. #EmbraceFlaws",
		"Reflecting on the past week, the small victories, and the lessons learned from the not-so-small failures. Growth is messy but necessary. #GrowthMindset",
	}
	buf := bytes.NewBuffer(nil)
	for k := range feed.Entries {
		title := TextConstruct{
			XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
			Text:    titles[k],
		}

		xml.EscapeText(buf, []byte(bodies[k]))
		tmp := make([]byte, buf.Len())
		copy(tmp, buf.Bytes())
		buf.Reset()
		content := Content{
			Type: "text",
			Body: tmp,
		}
		published := DateConstruct{
			XMLName: xml.Name{Space: atom_xmlns, Local: "published"},
			T:       fixed_date,
		}
		updated := DateConstruct{
			XMLName: xml.Name{Space: atom_xmlns, Local: "updated"},
			T:       fixed_date,
		}
		edited := DateConstruct{
			XMLName: xml.Name{Space: app_xmlns, Local: "edited"},
			T:       fixed_date,
		}
		id := URI{
			XMLName: xml.Name{Space: atom_xmlns, Local: "id"},
			Target:  uuids[k],
		}
		fixed_date = fixed_date.Add(-24 * time.Hour)
		feed.Entries[k] = &Entry{
			Id:         id,
			Authors:    []Person{author},
			Title:      title,
			Updated:    updated,
			Published:  &published,
			Rights:     &rights,
			Categories: []Category{category},
			Content:    content,
			Control:    &PublishingControl{Draft: "yes"},
			Edited:     &edited,
		}
	}

	cats := []Category{{
		Term: "banana",
	}, {
		Term: "apple",
	}, {
		Term: "orange",
	},
	}

	appcat := Categories{
		Fixed:      "yes",
		Scheme:     "https://example.org/scheme/",
		Categories: cats,
	}

	accept_1 := Accept{
		Text: "text/plain",
	}
	accept_2 := Accept{
		Text: "application/atom+xml;type=entry",
	}

	title := &TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
		Text:    "Blog Posts",
	}

	collection := &Collection{
		Href:       "https://example.org/submit/",
		Title:      title,
		Categories: []Categories{appcat},
		Accepts:    []Accept{accept_1, accept_2},
	}

	feed.Collection = collection

	bw := bufio.NewWriter(buf)
	if e := feed.MarshalTo(bw); e != nil {
		t.Fatal(e)
	} else if e := bw.Flush(); e != nil {
		t.Fatal(e)
	}
	dec := xml.NewDecoder(buf)
	for {
		if tok, e := dec.Token(); errors.Is(e, io.EOF) {
			break
		} else if e != nil {
			t.Fatal(e)
		} else if se, ok := tok.(xml.StartElement); ok {
			switch se.Name.Local {
			case "collection", "accept", "categories", "control", "draft", "edited":
				if se.Name.Space != app_xmlns {
					t.Fatal("unexpected xmlns")
				}
			default:
				if se.Name.Space != atom_xmlns {
					t.Fatal("unexpected xmlns")
				}
			}
		}
	}
}

func TestMarshalEntryWithSource(t *testing.T) {
	author := Person{
		XMLName: xml.Name{Space: atom_xmlns, Local: "author"},
		Name:    "John Doe",
		URI:     &URI{XMLName: xml.Name{Space: atom_xmlns, Local: "uri"}, Target: "mailto:johndoe@example.org"},
	}
	fixed_date := time.Date(2025, time.February, 9, 16, 20, 0, 0, time.Local)
	rights := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "rights"},
		Text:    "Copyright (c) 2025, John Doe",
	}
	feed_link := Link{
		Href:     "https://example.org/feed",
		Relation: "self",
		Type:     "application/atom+xml",
	}
	category := Category{
		Term: "blog posts",
	}
	feed_title := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
		Type:    "html",
		Text:    "John Doe Blog",
	}

	entry := func() *Entry {
		uuids := [1]string{
			"urn:uuid:59592fc2-0a7d-47fb-aceb-7d4a7bd6985a",
		}
		titles := [1]string{
			"Random Thoughts 3",
		}
		bodies := [1]string{
			"Today's musings on the absurdity of life, the universe, and why my coffee never tastes the same twice. #LifeQuestions #CoffeeConundrums",
		}

		buf := bytes.NewBuffer(nil)

		title := TextConstruct{
			XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
			Text:    titles[0],
		}

		xml.EscapeText(buf, []byte(bodies[0]))
		tmp := make([]byte, buf.Len())
		copy(tmp, buf.Bytes())
		buf.Reset()
		content := Content{
			Type: "text",
			Body: tmp,
		}
		published := DateConstruct{
			XMLName: xml.Name{Space: atom_xmlns, Local: "published"},
			T:       fixed_date,
		}
		updated := DateConstruct{
			XMLName: xml.Name{Space: atom_xmlns, Local: "updated"},
			T:       fixed_date,
		}
		edited := DateConstruct{
			XMLName: xml.Name{Space: app_xmlns, Local: "edited"},
			T:       fixed_date,
		}
		id := URI{
			XMLName: xml.Name{Space: atom_xmlns, Local: "id"},
			Target:  uuids[0],
		}
		return &Entry{
			Id:         id,
			Authors:    []Person{author},
			Title:      title,
			Updated:    updated,
			Published:  &published,
			Rights:     &rights,
			Categories: []Category{category},
			Content:    content,
			Control:    &PublishingControl{Draft: "yes"},
			Edited:     &edited,
		}
	}()
	cats := []Category{{
		Term: "banana",
	}, {
		Term: "apple",
	}, {
		Term: "orange",
	},
	}

	appcat := Categories{
		Fixed:      "yes",
		Scheme:     "https://example.org/scheme/",
		Categories: cats,
	}

	accept_1 := Accept{
		Text: "text/plain",
	}
	accept_2 := Accept{
		Text: "application/atom+xml;type=entry",
	}

	title := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
		Text:    "Blog Posts",
	}

	collection := &Collection{
		Href:       "https://example.org/submit/",
		Title:      &title,
		Categories: []Categories{appcat},
		Accepts:    []Accept{accept_1, accept_2},
	}

	source := &Source{
		Title:      &feed_title,
		Links:      []Link{feed_link},
		Categories: []Category{category},
		Collection: collection,
	}
	entry.Source = source

	buf := bytes.NewBuffer(nil)
	bw := bufio.NewWriter(buf)
	if e := entry.MarshalTo(bw, nil); e != nil {
		t.Fatal(e)
	} else if e := bw.Flush(); e != nil {
		t.Fatal(e)
	}

	dec := xml.NewDecoder(buf)
	for {
		if tok, e := dec.Token(); errors.Is(e, io.EOF) {
			break
		} else if e != nil {
			t.Fatal(e)
		} else if se, ok := tok.(xml.StartElement); ok {
			switch se.Name.Local {
			case "collection", "accept", "categories", "control", "draft", "edited":
				if se.Name.Space != app_xmlns {
					t.Fatal("unexpected xmlns")
				}
			default:
				if se.Name.Space != atom_xmlns {
					t.Fatal("unexpected xmlns")
				}
			}
		}
	}
}

func TestMarshalService(t *testing.T) {
	cats := []Category{{
		Term: "banana",
	}, {
		Term: "apple",
	}, {
		Term: "orange",
	},
	}

	appcat := Categories{
		Fixed:      "yes",
		Scheme:     "https://example.org/scheme/",
		Categories: cats,
	}

	accept_1 := Accept{
		Text: "text/plain",
	}
	accept_2 := Accept{
		Text: "application/atom+xml;type=entry",
	}

	title := TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
		Text:    "Blog Posts",
	}

	collection := &Collection{
		Href:       "https://example.org/submit/",
		Title:      &title,
		Categories: []Categories{appcat},
		Accepts:    []Accept{accept_1, accept_2},
	}

	workspace := Workspace{
		Title:       title,
		Collections: []*Collection{collection},
	}

	service := &Service{
		Workspaces: []Workspace{workspace},
	}

	buf := bytes.NewBuffer(nil)
	bw := bufio.NewWriter(buf)
	if e := service.MarshalTo(bw); e != nil {
		t.Fatal(e)
	} else if e := bw.Flush(); e != nil {
		t.Fatal(e)
	}
	dec := xml.NewDecoder(buf)
	for {
		if tok, e := dec.Token(); errors.Is(e, io.EOF) {
			break
		} else if e != nil {
			t.Fatal(e)
		} else if se, ok := tok.(xml.StartElement); ok {
			switch se.Name.Local {
			case "category", "title":
				if se.Name.Space != atom_xmlns {
					t.Fatal("unexpected xmlns")
				}
			default:
				if se.Name.Space != app_xmlns {
					t.Fatal("unexpected xmlns")
				}
			}
		}
	}
}
