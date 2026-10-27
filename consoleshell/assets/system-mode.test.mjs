import {readFileSync} from "node:fs";
import {runInNewContext} from "node:vm";
import assert from "node:assert/strict";
import test from "node:test";

const source = readFileSync(new URL("./shell.js", import.meta.url), "utf8");

function setup(options, stored = null, blocked = false) {
  const classes = new Set();
  const listeners = new Set();
  const media = {
    matches: false,
    addEventListener(name, listener) { assert.equal(name, "change"); listeners.add(listener); },
    removeEventListener(name, listener) { assert.equal(name, "change"); listeners.delete(listener); }
  };
  const root = {dataset: {}, classList: {
    contains(name) { return classes.has(name); },
    toggle(name, enabled) { enabled ? classes.add(name) : classes.delete(name); }
  }};
  const document = {documentElement: root, addEventListener() {}};
  let factory;
  const window = {Alpine: {data(name, fn) { assert.equal(name, "consoleShell"); factory = fn; }}, matchMedia() { return media; }};
  const localStorage = {getItem() { if (blocked) throw new Error("blocked"); return stored; }};
  runInNewContext(source, {window, document, localStorage});
  const component = factory(options);
  component.init();
  return {classes, listeners, component, change(matches) { media.matches = matches; for (const listener of listeners) listener(media); }};
}

test("System follows the OS and removes its listener on destruction", () => {
  const state = setup({persist:false, theme:"goshtoso", colorScheme:"system"});
  assert.equal(state.listeners.size, 1);
  state.change(true);
  assert.equal(state.classes.has("dark"), true);
  state.change(false);
  assert.equal(state.classes.has("dark"), false);
  state.component.destroy();
  assert.equal(state.listeners.size, 0);
});

test("forced modes do not subscribe to OS changes", () => {
  for (const colorScheme of ["light", "dark"]) {
    const state = setup({persist:false, colorScheme});
    assert.equal(state.listeners.size, 0);
  }
});

test("saved browser mode remains authoritative only when persistence is enabled", () => {
  assert.equal(setup({persist:true, colorScheme:"system"}, "false").listeners.size, 0);
  assert.equal(setup({persist:false, colorScheme:"system"}, "false").listeners.size, 1);
  assert.equal(setup({persist:true, colorScheme:"system"}, null, true).listeners.size, 1);
});
