// The colour theme control.
//
// Three states, because two is not enough: a reader can ask for light, ask
// for dark, or leave it to the system -- and the third is a real choice, not
// the absence of one. Someone who has picked dark by hand should keep it when
// their laptop switches at sunset; someone following the system should not.
//
// The theme itself is applied by an inline script in the document head, which
// has to happen before the first paint. This file is only the control: it
// marks the current state, records a new one, and keeps a system-following
// page in step when the system changes underneath it.
(function () {
  var KEY = 'tobar-segais.theme';
  var root = document.documentElement;
  var system = matchMedia('(prefers-color-scheme: dark)');

  function choice() {
    var c = root.dataset.themeChoice;
    return c === 'light' || c === 'dark' ? c : 'system';
  }

  function apply(c) {
    var dark = c === 'dark' || (c === 'system' && system.matches);
    root.dataset.theme = dark ? 'dark' : 'light';
    root.dataset.themeChoice = c;
    mark();
  }

  // The control is rebuilt with the rest of the page on in-place navigation,
  // so the marking is done by query rather than held in a variable.
  function mark() {
    var c = choice();
    document.querySelectorAll('.theme').forEach(function (group) {
      group.hidden = false; // it does nothing without this script; now it does
      group.querySelectorAll('button[value]').forEach(function (b) {
        b.setAttribute('aria-pressed', b.value === c ? 'true' : 'false');
      });
    });
  }

  document.addEventListener('click', function (ev) {
    var b = ev.target.closest ? ev.target.closest('.theme button[value]') : null;
    if (!b) return;
    ev.preventDefault();
    try {
      localStorage.setItem(KEY, b.value);
    } catch (e) {
      // Unwritable storage costs the reader the choice next time, not now.
    }
    apply(b.value);
  });

  // Only while following the system: an explicit choice outranks it.
  system.addEventListener('change', function () {
    if (choice() === 'system') apply('system');
  });

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', mark);
  } else {
    mark();
  }
})();
