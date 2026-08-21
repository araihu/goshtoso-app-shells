(function () {
  "use strict";

  var sidebarScrollTop = 0;
  var tocObserver = null;

  function drawerFocusables(sidebar) {
    if (!sidebar) return [];
    return Array.prototype.slice.call(sidebar.querySelectorAll(
      'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), summary, [tabindex]:not([tabindex="-1"])'
    )).filter(function (element) {
      return element.tabIndex >= 0 &&
        element.getAttribute("aria-hidden") !== "true" &&
        element.getClientRects().length > 0 &&
        getComputedStyle(element).visibility !== "hidden";
    });
  }

  function registerAlpineData() {
    if (!window.Alpine || window.__componentDocShellAlpineRegistered) return;
    window.__componentDocShellAlpineRegistered = true;
    window.Alpine.data("componentDocShell", function (options) {
      var persist = !!(options && options.persist);
      var persistTheme = !!(options && options.persistTheme);
      var root = document.documentElement;
      var configuredTheme = (options && options.theme) || "araihu";
      var theme = root.dataset.themeSource === "preference" ? root.getAttribute("data-theme") || configuredTheme : configuredTheme;
      var dark = root.classList.contains("dark");
      var sidebarMedia = window.matchMedia("(min-width: 1024px)");
      var syncSidebarPersistence = null;
      var containDrawerTab = null;
      var containDrawerFocus = null;
      var syncDrawerFocus = null;
      return {
        theme: theme,
        dark: dark,
        persist: persist,
        persistTheme: persistTheme,
        sidebarOpen: false,
        sidebarPersistent: sidebarMedia.matches,
        init: function () {
          var self = this;
          var sidebar = document.getElementById("componentdocshell-sidebar");
          document.documentElement.setAttribute("data-theme", self.theme);
          document.documentElement.classList.toggle("dark", self.dark);
          syncDrawerFocus = function () {
            if (!(self.sidebarOpen && !self.sidebarPersistent)) return;
            self.$nextTick(function () {
              if (!(self.sidebarOpen && !self.sidebarPersistent) || !sidebar || sidebar.contains(document.activeElement)) return;
              var focusables = drawerFocusables(sidebar);
              if (focusables.length) focusables[0].focus({ preventScroll: true });
            });
          };
          containDrawerTab = function (event) {
            if (event.key !== "Tab") return;
            if (!(self.sidebarOpen && !self.sidebarPersistent) || !sidebar) return;
            var focusables = drawerFocusables(sidebar);
            if (!focusables.length) {
              event.preventDefault();
              return;
            }
            var first = focusables[0];
            var last = focusables[focusables.length - 1];
            var active = document.activeElement;
            if (event.shiftKey && (active === first || !sidebar.contains(active))) {
              event.preventDefault();
              last.focus({ preventScroll: true });
            } else if (!event.shiftKey && (active === last || !sidebar.contains(active))) {
              event.preventDefault();
              first.focus({ preventScroll: true });
            }
          };
          containDrawerFocus = function (event) {
            if (!(self.sidebarOpen && !self.sidebarPersistent) || !sidebar || sidebar.contains(event.target)) return;
            var focusables = drawerFocusables(sidebar);
            if (focusables.length) focusables[0].focus({ preventScroll: true });
          };
          document.addEventListener("keydown", containDrawerTab, true);
          document.addEventListener("focusin", containDrawerFocus, true);
          syncSidebarPersistence = function (event) {
            self.sidebarPersistent = event.matches;
            if (event.matches) self.sidebarOpen = false;
            syncDrawerFocus();
          };
          if (sidebarMedia.addEventListener) sidebarMedia.addEventListener("change", syncSidebarPersistence);
          else if (sidebarMedia.addListener) sidebarMedia.addListener(syncSidebarPersistence);
          self.$watch("sidebarOpen", syncDrawerFocus);
          self.$watch("sidebarPersistent", syncDrawerFocus);
          self.$watch("theme", function (value) {
            document.documentElement.dataset.themeSource = "preference";
            document.documentElement.setAttribute("data-theme", value);
            if (!self.persistTheme) return;
            try { localStorage.setItem("theme", value); } catch (_) {}
          });
        },
        destroy: function () {
          if (syncSidebarPersistence && sidebarMedia.removeEventListener) sidebarMedia.removeEventListener("change", syncSidebarPersistence);
          else if (syncSidebarPersistence && sidebarMedia.removeListener) sidebarMedia.removeListener(syncSidebarPersistence);
          if (containDrawerTab) document.removeEventListener("keydown", containDrawerTab, true);
          if (containDrawerFocus) document.removeEventListener("focusin", containDrawerFocus, true);
          syncSidebarPersistence = null;
          containDrawerTab = null;
          containDrawerFocus = null;
          syncDrawerFocus = null;
        },
        setTheme: function (value) {
          document.documentElement.dataset.themeSource = "preference";
          this.theme = value;
        },
        toggleDark: function () {
          this.dark = !this.dark;
          document.documentElement.classList.toggle("dark", this.dark);
          if (!this.persist) return;
          try { localStorage.setItem("darkMode", String(this.dark)); } catch (_) {}
        }
      };
    });
  }

  if (window.Alpine) registerAlpineData();
  document.addEventListener("alpine:init", registerAlpineData, { once: true });

  function mainContent() {
    return document.getElementById("main-content");
  }

  function mainFocusTarget() {
    var main = mainContent();
    if (!main) return null;
    var heading = main.querySelector("h1");
    var target = heading || main;
    if (!target.hasAttribute("tabindex")) target.setAttribute("tabindex", "-1");
    return target;
  }

  function focusMain() {
    var target = mainFocusTarget();
    if (!target) return;
    target.focus({ preventScroll: true });
  }

  function syncFamilySelect() {
    var menu = document.getElementById("componentdocshell-family-select-control");
    var active = document.querySelector('.component-doc-shell__family-links [aria-current="location"]');
    var href = active && active.getAttribute("href");
    if (!menu || !href || !window.Alpine) return;
    var menuState = window.Alpine.$data(menu);
    if (menuState) {
      menuState.initialFamilyHref = href;
      menuState.familyHref = href;
    }
    var select = menu.querySelector("[data-select-config]");
    var selectState = select && window.Alpine.$data(select);
    if (selectState && typeof selectState.syncFromInput === "function") selectState.syncFromInput(href);
  }

  function closeFamilySelect() {
    var select = document.querySelector("#componentdocshell-family-select-control [data-select-config]");
    if (!select || !window.Alpine) return;
    var state = window.Alpine.$data(select);
    if (!state) return;
    state.isOpen = false;
    state.openedWithKeyboard = false;
  }

  function scrollTarget(target, behavior) {
    var scroller = document.getElementById("page-scroll");
    if (!target || !scroller) return;
    document.documentElement.scrollTop = 0;
    document.body.scrollTop = 0;
    var margin = parseFloat(getComputedStyle(target).scrollMarginTop) || 0;
    var nextTop = scroller.scrollTop + target.getBoundingClientRect().top - scroller.getBoundingClientRect().top - margin;
    nextTop = Math.max(0, nextTop);
    scroller.scrollTo({ top: nextTop, behavior: behavior || "auto" });
  }

  function tocScrollBehavior() {
    return window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth";
  }

  function buildTOC() {
    var rail = document.querySelector("[data-componentdocshell-toc]");
    var list = document.querySelector("[data-componentdocshell-toc-list]");
    var content = mainContent();
    if (!rail || !list || !content || rail.dataset.enabled !== "true") return;
    if (tocObserver) tocObserver.disconnect();
    var headings = Array.prototype.slice.call(content.querySelectorAll("[data-toc-heading][id]"));
    list.replaceChildren();
    rail.hidden = headings.length < 2;
    if (headings.length < 2) return;
    headings.forEach(function (heading) {
      var link = document.createElement("a");
      link.href = "#" + heading.id;
      link.textContent = (heading.textContent || "").trim();
      link.setAttribute("data-toc-link", heading.id);
      link.addEventListener("click", function (event) {
        event.preventDefault();
        history.replaceState(null, "", "#" + heading.id);
        scrollTarget(heading, tocScrollBehavior());
      });
      list.appendChild(link);
    });
    var hashID = decodeURIComponent((window.location.hash || "").replace(/^#/, ""));
    var active = headings.find(function (heading) { return heading.id === hashID; });
    if (active) requestAnimationFrame(function () { scrollTarget(active, "auto"); });
    if (!("IntersectionObserver" in window)) return;
    tocObserver = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (!entry.isIntersecting) return;
        list.querySelectorAll("a").forEach(function (link) {
          link.classList.toggle("is-active", link.getAttribute("href") === "#" + entry.target.id);
        });
      });
    }, { root: document.getElementById("page-scroll"), rootMargin: "0px 0px -70%", threshold: 0.1 });
    headings.forEach(function (heading) { tocObserver.observe(heading); });
  }

  document.addEventListener("htmx:beforeSwap", function () {
    var sidebar = document.querySelector(".sidebar-scroll");
    if (sidebar) sidebarScrollTop = sidebar.scrollTop;
  });

  document.addEventListener("htmx:afterSwap", function (event) {
    if (!event.detail || !event.detail.target || event.detail.target.id !== "main-content") return;
    var sidebar = document.querySelector(".sidebar-scroll");
    if (sidebar) sidebar.scrollTop = sidebarScrollTop;
    var pageScroll = document.getElementById("page-scroll");
    if (pageScroll) pageScroll.scrollTo({ top: 0 });
    syncFamilySelect();
    window.dispatchEvent(new CustomEvent("componentdocshell:navigated"));
    buildTOC();
    focusMain();
  });

  document.addEventListener("htmx:historyRestore", function () {
    if (!mainContent()) return;
    syncFamilySelect();
    window.dispatchEvent(new CustomEvent("componentdocshell:navigated"));
    buildTOC();
    focusMain();
  });

  window.addEventListener("componentdocshell:close-family-select", closeFamilySelect);
  window.componentDocShell = { buildTOC: buildTOC, focusMain: focusMain, syncFamilySelect: syncFamilySelect, closeFamilySelect: closeFamilySelect };
  document.addEventListener("DOMContentLoaded", function () {
    mainFocusTarget();
    buildTOC();
  });
})();
