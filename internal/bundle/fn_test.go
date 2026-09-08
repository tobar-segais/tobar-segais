package bundle

import "testing"

func TestFilenameInference(t *testing.T) {
	cases := []string{
		"foo-1.0.0.zip", "acme-docs-4.2.1.zip", "handbook-2.0.zip",
		"foo-1.0.0-SNAPSHOT.zip", "foo-v1.0.0.zip", "foo_1.0.0.zip",
		"my.product-3.1.4.zip", "handbook.zip", "foo-1.0.0-rc1.jar",
	}
	for _, name := range cases {
		p := makeZip(t, name, map[string]string{"nav.html": "<html><body><ul><li><a href=\"x.html\">X</a></li></ul></body></html>"})
		id := identify(t, p)
		t.Logf("%-26s -> slug=%-18s version=%-16s (%s)", name, id.Slug, id.Version, id.Source)
	}
}

// Nothing has to be declared, and a partial declaration is not discarded.
func TestIdentityIsAssembledFromWhateverIsAvailable(t *testing.T) {
	nav := func(metas string) string {
		return `<html><head>` + metas + `</head><body><ul>` +
			`<li><a href="x.html">X</a></li></ul></body></html>`
	}
	cases := []struct {
		name, metas       string
		wantSlug, wantVer string
		wantSource        string
	}{
		{"foo-1.0.0.zip", "", "foo", "1.0.0", "filename"},
		{"foo-1.0.0.zip", `<meta name="tobar-segais.slug" content="bar">`,
			"bar", "1.0.0", "nav+filename"},
		{"foo-1.0.0.zip", `<meta name="tobar-segais.version" content="9.9.9">`,
			"foo", "9.9.9", "nav+filename"},
		{"foo-1.0.0.zip",
			`<meta name="tobar-segais.slug" content="bar"><meta name="tobar-segais.version" content="9.9.9">`,
			"bar", "9.9.9", "nav"},
	}
	for _, c := range cases {
		p := makeZip(t, c.name, map[string]string{"nav.html": nav(c.metas)})
		id := identify(t, p)
		if id.Slug != c.wantSlug || id.Version != c.wantVer || id.Source != c.wantSource {
			t.Errorf("%s with %q:\n got  slug=%s version=%s source=%s\n want slug=%s version=%s source=%s",
				c.name, c.metas, id.Slug, id.Version, id.Source, c.wantSlug, c.wantVer, c.wantSource)
		}
	}
}

// A bundle with no version anywhere is served as "latest" rather than at a
// URL with an empty path segment.
func TestUnversionedBundleBecomesLatest(t *testing.T) {
	p := makeZip(t, "handbook.zip", map[string]string{
		"nav.html": `<html><body><ul><li><a href="x.html">X</a></li></ul></body></html>`,
	})
	id := identify(t, p)
	if id.Slug != "handbook" {
		t.Errorf("slug = %q", id.Slug)
	}
	if id.Version != LatestVersion {
		t.Errorf("version = %q, want %q", id.Version, LatestVersion)
	}
	if id.Source != "filename+default" {
		t.Errorf("source = %q", id.Source)
	}
}
