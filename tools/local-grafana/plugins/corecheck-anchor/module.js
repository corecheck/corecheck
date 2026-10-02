define(["react", "@grafana/data", "@grafana/runtime"], function (reactNs, dataNs, runtimeNs) {
  var React = reactNs && reactNs.useEffect ? reactNs : reactNs.default || reactNs;
  var data = dataNs && dataNs.PanelPlugin ? dataNs : dataNs.default || dataNs;
  var runtime = runtimeNs && runtimeNs.locationService ? runtimeNs : runtimeNs.default || runtimeNs;

  var sections = {
    "functional-trend": {variable: "functional_test", title: "Functional duration"},
    "unit-trend": {variable: "unit_test", title: "Unit duration"},
  };

  function currentValue(variable) {
    var current = variable && variable.current;
    var value = current && current.value;
    if (value == null && variable && variable.state) {
      value = variable.state.value;
    }
    if (Array.isArray(value)) {
      return value.length ? String(value[0]) : "";
    }
    return value == null ? "" : String(value);
  }

  function variableByName(name) {
    var list = runtime.getTemplateSrv().getVariables() || [];
    for (var i = 0; i < list.length; i++) {
      if (list[i] && list[i].name === name) {
        return list[i];
      }
    }
    return null;
  }

  function parsedHash() {
    var raw = (window.location.hash || "").replace(/^#/, "");
    try {
      raw = decodeURIComponent(raw);
    } catch (err) {}
    var idx = raw.indexOf("~");
    if (idx === -1) {
      return {key: raw, value: ""};
    }
    return {key: raw.slice(0, idx), value: raw.slice(idx + 1)};
  }

  function findTitle(prefix) {
    var nodes = document.body.getElementsByTagName("*");
    for (var i = 0; i < nodes.length; i++) {
      var el = nodes[i];
      if (el.childElementCount > 0) {
        continue;
      }
      var text = (el.textContent || "").trim();
      if (text.indexOf(prefix) === 0) {
        return el;
      }
    }
    return null;
  }

  function dashboardScroller() {
    var nodes = document.body.getElementsByTagName("*");
    var best = null;
    var bestHeight = 0;
    for (var i = 0; i < nodes.length; i++) {
      var el = nodes[i];
      var style = window.getComputedStyle(el);
      if (!/(auto|scroll)/.test(style.overflowY)) {
        continue;
      }
      if (el.scrollHeight > el.clientHeight + 50 && el.clientHeight > 200 && el.scrollHeight > bestHeight) {
        best = el;
        bestHeight = el.scrollHeight;
      }
    }
    return best;
  }

  function scrollToTitle(prefix) {
    var el = findTitle(prefix);
    if (el) {
      el.scrollIntoView({behavior: "auto", block: "start"});
      return true;
    }
    var scroller = dashboardScroller();
    if (scroller) {
      scroller.scrollTop += Math.round(scroller.clientHeight * 0.75);
    }
    return false;
  }

  function Anchor() {
    React.useEffect(function () {
      function applyDefault(name) {
        var variable = variableByName(name);
        if (!variable) {
          return false;
        }
        if (currentValue(variable)) {
          return true;
        }
        var options = variable.options || [];
        if (!options.length || options[0].value == null || options[0].value === "") {
          return false;
        }
        var query = {};
        query["var-" + name] = options[0].value;
        runtime.locationService.partial(query, true);
        return true;
      }

      function followHash() {
        var parsed = parsedHash();
        var section = sections[parsed.key];
        if (!section) {
          return;
        }
        if (parsed.value) {
          var query = {};
          query["var-" + section.variable] = parsed.value;
          runtime.locationService.partial(query, true);
        }
        var delays = [0, 100, 300, 600, 1000, 1500];
        for (var i = 0; i < delays.length; i++) {
          setTimeout(function () {
            scrollToTitle(section.title);
          }, delays[i]);
        }
      }

      var attempts = 0;
      var timer = setInterval(function () {
        attempts += 1;
        var functionalReady = applyDefault("functional_test");
        var unitReady = applyDefault("unit_test");
        if ((functionalReady && unitReady) || attempts > 40) {
          clearInterval(timer);
        }
      }, 500);

      followHash();
      window.addEventListener("hashchange", followHash);
      return function () {
        clearInterval(timer);
        window.removeEventListener("hashchange", followHash);
      };
    }, []);

    return React.createElement("div", {id: "tests-nav"});
  }

  return {
    plugin: new data.PanelPlugin(Anchor),
  };
});
