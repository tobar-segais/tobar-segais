// In-place navigation.
//
// Following a link fetches the target and swaps the page shell, keeping the
// sidebar scrolled where the reader left it. A full page load would throw
// that away.
//
// Every pane is replaced, not just the content: the sidebar carries the
// current-page mark, the version pills and the page tree, and all three
// differ between pages, while the footer carries the bundle's own copyright
// notice. Keeping the old ones leaves them lying -- two versions showing as
// current, the tree of the version you just left, or one manual's notice
// under another's pages. The sidebar's scroll position is restored
// afterwards, which is the state worth keeping.
//
// The page is fetched before anything changes, so a link that turns out to be
// missing shows the not-found page in the content pane rather than replacing
// the site around it. Nothing here is required: with JavaScript off, or if a
// fetch fails, links are ordinary links and the browser handles them.
(function () {
  var script = document.currentScript;
  var site = script.getAttribute('data-site') || '';

  var MAIN = 'main.doc-shell';
  var SIDEBAR = 'aside.nav';
  var FOOTER = 'footer.footer';

  function isPage(url) {
    if (url.origin !== location.origin) return false;
    var p = url.pathname;
    if (site && p.indexOf(site) !== 0) return false;
    if (/\.(png|jpe?g|gif|svg|pdf|zip|jar|css|js|json|txt|ico)$/i.test(p)) return false;
    return true;
  }

  // Bring the current page into view when it is outside the visible part of
  // the sidebar. A long tree can easily put it off-screen, and a reader
  // should not have to hunt for where they are.
  function revealCurrent() {
    var nav = document.querySelector(SIDEBAR);
    if (!nav) return;
    var current = nav.querySelector('[aria-current="page"]');
    if (!current) return;

    var navBox = nav.getBoundingClientRect();
    var itemBox = current.getBoundingClientRect();
    if (itemBox.top >= navBox.top && itemBox.bottom <= navBox.bottom) return;

    // Centre it rather than putting it hard against an edge, so the
    // surrounding pages stay visible for context.
    nav.scrollTop += (itemBox.top - navBox.top) - (nav.clientHeight / 2) + (itemBox.height / 2);
  }

  function swap(html, url, push) {
    var doc = new DOMParser().parseFromString(html, 'text/html');
    var nextMain = doc.querySelector(MAIN);
    var hereMain = document.querySelector(MAIN);
    if (!nextMain || !hereMain) {
      location.href = url.href; // not a page we understand; hand over
      return;
    }

    var hereNav = document.querySelector(SIDEBAR);
    var nextNav = doc.querySelector(SIDEBAR);
    var scroll = hereNav ? hereNav.scrollTop : 0;

    hereMain.replaceWith(nextMain);
    if (hereNav && nextNav) {
      hereNav.replaceWith(nextNav);
      nextNav.scrollTop = scroll;
    } else if (hereNav !== null && nextNav === null) {
      hereNav.remove();
    } else if (hereNav === null && nextNav !== null) {
      var body = document.querySelector('.body');
      if (body) body.insertBefore(nextNav, body.firstChild);
    }

    // The footer is not decoration: a bundle can carry its own copyright
    // notice, so moving between bundles changes it. Leaving it behind would
    // show one manual's notice under another's pages.
    var hereFoot = document.querySelector(FOOTER);
    var nextFoot = doc.querySelector(FOOTER);
    if (hereFoot && nextFoot) hereFoot.replaceWith(nextFoot);

    document.title = doc.title;
    if (push) history.pushState({}, '', url.href);

    if (url.hash) {
      var target = document.getElementById(url.hash.slice(1));
      if (target) {
        target.scrollIntoView();
        revealCurrent();
        return;
      }
    }
    window.scrollTo(0, 0);
    revealCurrent();
  }

  function go(url, push) {
    var main = document.querySelector(MAIN);
    if (main) main.setAttribute('aria-busy', 'true');
    fetch(url.href, { headers: { 'Accept': 'text/html' } })
      .then(function (r) { return r.text(); })   // a 404 is still a page
      .then(function (html) { swap(html, url, push); })
      .catch(function () { location.href = url.href; })
      .finally(function () {
        var m = document.querySelector(MAIN);
        if (m) m.removeAttribute('aria-busy');
      });
  }

  document.addEventListener('click', function (ev) {
    if (ev.defaultPrevented || ev.button !== 0) return;
    if (ev.metaKey || ev.ctrlKey || ev.shiftKey || ev.altKey) return;

    var a = ev.target.closest ? ev.target.closest('a[href]') : null;
    if (!a || a.hasAttribute('download') || a.getAttribute('target')) return;

    var url;
    try {
      url = new URL(a.href, location.href);
    } catch (e) {
      return;
    }
    if (!isPage(url)) return;

    // A link to the same page with only a fragment is the browser's job.
    if (url.pathname === location.pathname && url.hash) return;

    ev.preventDefault();
    go(url, true);
  });

  // The version dropdown carries an inline handler so it still works with
  // this script absent. Intercept it during capture, before that handler
  // runs, so switching version swaps in place like any other link.
  document.addEventListener('change', function (ev) {
    var el = ev.target;
    if (!el || el.tagName !== 'SELECT' || !el.hasAttribute('data-nav')) return;
    var url;
    try {
      url = new URL(el.value, location.href);
    } catch (e) {
      return;
    }
    if (!isPage(url)) return;
    ev.stopPropagation();
    ev.preventDefault();
    go(url, true);
  }, true);

  window.addEventListener('popstate', function () {
    go(new URL(location.href), false);
  });

  // Also on a normal page load, where the sidebar may open below the fold.
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', revealCurrent);
  } else {
    revealCurrent();
  }
})();
