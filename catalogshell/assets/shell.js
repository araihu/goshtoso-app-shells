(function () {
  "use strict";

  var sidebarScrollTop = 0;
  var tocObserver = null;

  function mainContent() {
    return document.getElementById("main-content");
  }

  function focusMain() {
    var main = mainContent();
    if (!main) return;
    var heading = main.querySelector("h1");
    var target = heading || main;
    if (!target.hasAttribute("tabindex")) target.setAttribute("tabindex", "-1");
    target.focus({ preventScroll: true });
  }

  function buildTOC() {
    var rail = document.getElementById("catalogshell-toc");
    var list = document.getElementById("catalogshell-toc-list");
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
      list.appendChild(link);
    });
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
    buildTOC();
    focusMain();
  });

  window.catalogShell = { buildTOC: buildTOC, focusMain: focusMain };
  document.addEventListener("DOMContentLoaded", buildTOC);
})();
