(function () {
  "use strict";
  var sidebarScrollTop = 0;

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
      var persist = !!(o && o.persist), root = document.documentElement, configuredTheme = (o && o.theme) || "goshtoso";
      var theme = root.dataset.themeSource === "preference" ? root.dataset.theme || configuredTheme : configuredTheme;
      var dark = root.classList.contains("dark");
      return { sidebarOpen: false, init: function () { document.documentElement.dataset.theme=theme; document.documentElement.classList.toggle("dark",dark); }, closeDrawer: function (restoreFocus) { if (!this.sidebarOpen) return; this.sidebarOpen=false; if (restoreFocus === false) return; this.$nextTick(function () { var trigger = document.getElementById("consoleshell-menu"); if (trigger) trigger.focus(); }.bind(this)); } };
    });
  }
  function beforeSwap() { var sidebar=document.getElementById("consoleshell-sidebar-scroll"); if (sidebar) sidebarScrollTop=sidebar.scrollTop; }
  function afterSwap(event) {
    var target=event.detail && event.detail.task && event.detail.task.target;
    if (!target || (!target.matches("main.console-shell__main") && target !== document.body)) return;
    var sidebar=document.getElementById("consoleshell-sidebar-scroll"); if (sidebar) sidebar.scrollTop=sidebarScrollTop;
    var main=target === document.body ? document.querySelector("main.console-shell__main") : document.getElementById(target.id);
    if (!main) return;
    reconcileNavigation(main); main.scrollTo({top:0}); focusMain(main);
    window.dispatchEvent(new CustomEvent("consoleshell:navigated"));
  }
  function installLifecycle() {
    if (window.__consoleShellLifecycleInstalled) return;
    window.__consoleShellLifecycleInstalled = true;
    document.addEventListener("htmx:before:swap", beforeSwap);
    document.addEventListener("htmx:after:settle", afterSwap);
  }
  window.consoleShell = { focusMain: focusMain, reconcileNavigation: reconcileNavigation, installLifecycle: installLifecycle };
  if (window.Alpine) registerAlpineData(); else document.addEventListener("alpine:init", registerAlpineData, {once:true});
  installLifecycle();
})();
