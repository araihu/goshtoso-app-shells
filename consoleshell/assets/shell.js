(function () {
  "use strict";
  var sidebarScrollTop = 0;
  var listenersInstalled = false;

  function focusMain(main) {
    if (!main) return;
    var target = main.querySelector("[data-autofocus]") || main.querySelector("h1") || main;
    if (!target.hasAttribute("tabindex")) target.setAttribute("tabindex", "-1");
    target.focus({ preventScroll: true });
  }
  function reconcileNavigation(main) {
    var active = main && main.getAttribute("data-active-navigation");
    if (!active) return;
    document.querySelectorAll("[data-consoleshell-nav-id]").forEach(function (link) {
      if (link.getAttribute("data-consoleshell-nav-id") === active) link.setAttribute("aria-current", "page");
      else link.removeAttribute("aria-current");
    });
  }
  function registerAlpineData() {
    if (!window.Alpine || window.__consoleShellAlpineRegistered) return;
    window.__consoleShellAlpineRegistered = true;
    window.Alpine.data("consoleShell", function (o) {
      var persist = !!(o && o.persist), theme = (o && o.theme) || "goshtoso";
      var dark = o && (o.colorScheme === "dark" || (o.colorScheme === "system" && matchMedia("(prefers-color-scheme: dark)").matches));
      try { if (persist) { theme = localStorage.getItem("goshtoso-theme") || theme; var saved = localStorage.getItem("goshtoso-dark"); if (saved !== null) dark = saved === "true"; } } catch (_) {}
      return { sidebarOpen: false, init: function () { document.documentElement.dataset.theme=theme; document.documentElement.classList.toggle("dark",dark); }, closeDrawer: function (restoreFocus) { this.sidebarOpen=false; if (restoreFocus === false) return; this.$nextTick(function () { if (this.$refs.sidebarTrigger) this.$refs.sidebarTrigger.focus(); }.bind(this)); } };
    });
  }
  function beforeSwap() { var sidebar=document.getElementById("consoleshell-sidebar-scroll"); if (sidebar) sidebarScrollTop=sidebar.scrollTop; }
  function afterSwap(event) {
    var target=event.detail && event.detail.target;
    if (!target || !target.matches("main.console-shell__main")) return;
    var sidebar=document.getElementById("consoleshell-sidebar-scroll"); if (sidebar) sidebar.scrollTop=sidebarScrollTop;
    var main=document.getElementById(target.id) || document.querySelector("main.console-shell__main");
    if (!main) return;
    reconcileNavigation(main); main.scrollTo({top:0}); focusMain(main);
    window.dispatchEvent(new CustomEvent("consoleshell:navigated"));
  }
  function installLifecycle() {
    if (listenersInstalled) return; listenersInstalled=true;
    document.addEventListener("htmx:beforeSwap", beforeSwap);
    document.addEventListener("htmx:afterSettle", afterSwap);
    document.addEventListener("htmx:historyRestore", function () { var main=document.querySelector("main.console-shell__main"); if (!main) return; reconcileNavigation(main); main.scrollTo({top:0}); focusMain(main); window.dispatchEvent(new CustomEvent("consoleshell:navigated")); });
  }
  window.consoleShell = { focusMain: focusMain, reconcileNavigation: reconcileNavigation, installLifecycle: installLifecycle };
  if (window.Alpine) registerAlpineData(); else document.addEventListener("alpine:init", registerAlpineData, {once:true});
  installLifecycle();
})();
