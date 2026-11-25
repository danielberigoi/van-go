{
  const SCOOP = "[data-cone-scoop]";
  const PROPS = new Set(["value", "checked", "selected", "indeterminate"]);
  const registry = {};

  const compile = (names, src, isHandler) => {
    try {
      return isHandler
        ? new Function(...names, "event", src)
        : new Function(...names, `return (${src});`);
    } catch (e) {
      console.error(`cone: cannot compile ${JSON.stringify(src)}`, e);
      return () => {};
    }
  };

  const setAttr = (el, name, v) => {
    if (PROPS.has(name)) el[name] = name === "value" ? (v ?? "") : !!v;
    else if (v === false || v == null) el.removeAttribute(name);
    else el.setAttribute(name, v === true ? "" : v);
  };

  const attrText = v => v === false || v == null ? null : v === true ? "" : String(v);

  const mismatch = (el, src, server, browser) => {
    if (server !== browser) {
      console.warn(`cone: server rendered ${JSON.stringify(server)} but ${JSON.stringify(src)} gives ${JSON.stringify(browser)}`, el);
    }
  };

  const hydrate = root => {
    if (root.__cone) return root.__cone;
    const json = root.querySelector(":scope > script[data-cone-state]");
    const initial = json ? JSON.parse(json.textContent) : {};
    const names = Object.keys(initial);
    const states = Object.fromEntries(names.map(n => [n, reactive.state(initial[n])]));
    const args = names.map(n => states[n]);
    root.__cone = states;

    const els = [...root.querySelectorAll("*")].filter(el => el.closest(SCOOP) === root);
    for (const el of els) {
      for (const name of el.getAttributeNames()) {
        const value = el.getAttribute(name);
        if (name === "data-cone-text") {
          const f = compile(names, value);
          let first = true;
          reactive.derive(() => {
            const text = String(f(...args) ?? "");
            if (first) { first = false; mismatch(el, value, el.textContent, text); }
            el.textContent = text;
          });
        } else if (name.startsWith("data-cone-attr-")) {
          const attr = name.slice("data-cone-attr-".length), f = compile(names, value);
          let first = true;
          reactive.derive(() => {
            const v = f(...args);
            if (first) { first = false; mismatch(el, value, el.getAttribute(attr), attrText(v)); }
            setAttr(el, attr, v);
          });
        } else if (name.startsWith("data-cone-on-")) {
          const f = compile(names, value, true);
          el.addEventListener(name.slice("data-cone-on-".length), e => f(...args, e));
        }
      }
    }

    const hook = registry[root.dataset.coneScoop];
    if (hook) hook(root, states, reactive);
    return states;
  };

  const hydrateAll = (node = document) => {
    if (node.matches?.(SCOOP)) hydrate(node);
    node.querySelectorAll(SCOOP).forEach(hydrate);
  };

  const register = (name, fn) => {
    registry[name] = fn;
    document.querySelectorAll(SCOOP).forEach(root => {
      if (root.__cone && root.dataset.coneScoop === name) fn(root, root.__cone, reactive);
    });
  };

  window.cone = { hydrate, hydrateAll, register };
  hydrateAll();
}
