(() => {
  const state = {
    mode: "year",
    birthday: "1995-01-01",
    lifespan: 90,
    year: new Date().getFullYear(),
    startDate: new Date().toISOString().slice(0, 10),
    endDate: new Date(Date.now() + 90 * 864e5).toISOString().slice(0, 10),
    customTitle: "",
    fg: "#ffffff",
    fgOther: "#383838",
    bg: "#000000",
    grouping: "full",
    weekStart: "monday",
    dotShape: "circle",
    shapeStyle: "filled",
    fillFocus: "past",
    highlightToday: true,
    highlightColor: "#ff6b4a",
    dotGap: 12,
    dotSize: 26,
    showNumbers: false,
    showBoth: false,
    bgType: "match",
    bgUrl: "",
    bgRotate: 0,
    bgBlur: 0,
    bgDim: 30,
    bgFocalX: 50,
    bgFocalY: 50,
    gridOpacity: 100,
    showCard: false,
    cardStyle: "flat",
    cardOpacity: 35,
    cardRadius: 28,
    cardPad: 28,
    cardFade: 0,
    showAxes: false,
    showTitle: true,
    showLabels: false,
    showBranding: false,
    showProgress: true,
    muteText: false,
    showFooter: true,
    titleSize: 84,
    footerSize: 48,
    deviceId: "iphone-16-pro",
    width: 1206,
    height: 2622,
    widgetLayout: "none",
    margin: 48,
    marginTop: 48,
    statsScale: 100,
    habits: [
      { id: "h1", name: "Move", icon: "dumbbell", completions: {} },
      { id: "h2", name: "Read", icon: "book-open", completions: {} },
    ],
  };

  const STORAGE_KEY = "life-dots-config";
  const defaults = structuredClone(state);

  const $ = (sel) => document.querySelector(sel);
  const preview = $("#preview");
  const loupe = $("#loupe");
  const loupeImg = $("#loupeImg");
  const urlBox = $("#wallpaperUrl");
  const qrImg = $("#qrImg");
  const toast = $("#toast");

  function saveToLocalStorage() {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(toConfig(state)));
    } catch {
      /* quota / private mode */
    }
  }

  function loadFromLocalStorage() {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return null;
      return JSON.parse(raw);
    } catch {
      return null;
    }
  }

  function shapeForAPI(shape) {
    return shape === "square-nogap" ? "square" : shape;
  }

  function noGapForAPI(shape) {
    return shape === "square-nogap";
  }

  function toConfig(cfg) {
    return {
      mode: cfg.mode,
      birthday: cfg.birthday,
      lifespan: cfg.lifespan,
      year: cfg.year || new Date().getFullYear(),
      startDate: cfg.startDate,
      endDate: cfg.endDate,
      customTitle: cfg.customTitle,
      fg: cfg.fg,
      fgOther: cfg.fgOther,
      bg: cfg.bg,
      highlightColor: cfg.highlightColor,
      grouping: cfg.grouping,
      weekStart: cfg.weekStart,
      dotShape: shapeForAPI(cfg.dotShape),
      shapeStyle: cfg.shapeStyle,
      noGap: noGapForAPI(cfg.dotShape),
      dotGap: cfg.dotGap,
      dotSize: cfg.dotSize,
      fillFocus: cfg.fillFocus || "past",
      highlightToday: cfg.highlightToday,
      showNumbers: cfg.showNumbers,
      showBoth: cfg.showBoth,
      bgType: cfg.bgType,
      bgUrl: cfg.bgUrl,
      bgRotate: Number(cfg.bgRotate) || 0,
      bgBlur: cfg.bgBlur,
      bgDim: cfg.bgDim,
      bgFocalX: cfg.bgFocalX,
      bgFocalY: cfg.bgFocalY,
      gridOpacity: cfg.gridOpacity,
      showCard: cfg.showCard,
      cardStyle: "flat",
      cardOpacity: cfg.cardOpacity,
      cardRadius: cfg.cardRadius,
      cardPadH: cfg.cardPad,
      cardPadV: cfg.cardPad,
      cardFade: cfg.cardFade,
      showAxes: cfg.showAxes,
      showTitle: cfg.showTitle,
      showLabels: cfg.showLabels,
      showBranding: cfg.showBranding,
      showProgress: cfg.showProgress,
      muteText: !!cfg.muteText,
      showFooter: cfg.showFooter,
      titleSize: cfg.titleSize,
      footerSize: cfg.footerSize,
      deviceId: cfg.deviceId,
      width: cfg.width,
      height: cfg.height,
      widgetLayout: cfg.widgetLayout,
      margin: cfg.margin,
      marginTop: cfg.marginTop,
      statsScale: cfg.statsScale,
      habits: cfg.habits,
    };
  }

  function applyLoaded(cfg) {
    if (!cfg || typeof cfg !== "object") return;
    const map = {
      mode: "mode",
      birthday: "birthday",
      lifespan: "lifespan",
      year: "year",
      startDate: "startDate",
      endDate: "endDate",
      customTitle: "customTitle",
      fg: "fg",
      fgOther: "fgOther",
      bg: "bg",
      highlightColor: "highlightColor",
      grouping: "grouping",
      weekStart: "weekStart",
      dotShape: "dotShape",
      shapeStyle: "shapeStyle",
      fillFocus: "fillFocus",
      highlightToday: "highlightToday",
      bgType: "bgType",
      bgUrl: "bgUrl",
      bgRotate: "bgRotate",
      bgBlur: "bgBlur",
      bgDim: "bgDim",
      bgFocalX: "bgFocalX",
      bgFocalY: "bgFocalY",
      gridOpacity: "gridOpacity",
      showCard: "showCard",
      cardStyle: "cardStyle",
      cardOpacity: "cardOpacity",
      cardRadius: "cardRadius",
      cardPad: "cardPad",
      cardFade: "cardFade",
      showAxes: "showAxes",
      showTitle: "showTitle",
      showLabels: "showLabels",
      showBranding: "showBranding",
      showProgress: "showProgress",
      muteText: "muteText",
      showFooter: "showFooter",
      titleSize: "titleSize",
      footerSize: "footerSize",
      deviceId: "deviceId",
      width: "width",
      height: "height",
      widgetLayout: "widgetLayout",
      margin: "margin",
      marginTop: "marginTop",
      dotGap: "dotGap",
      dotSize: "dotSize",
      habits: "habits",
    };
    for (const [from, to] of Object.entries(map)) {
      if (cfg[from] !== undefined && cfg[from] !== null && cfg[from] !== "") {
        state[to] = from === "habits" ? structuredClone(cfg[from]) : cfg[from];
      }
    }
    if (cfg.showFooter === undefined && cfg.footerMode != null && cfg.footerMode !== "") {
      state.showFooter = cfg.footerMode !== "none";
    }
    if (cfg.noGap && (cfg.dotShape === "square" || !cfg.dotShape)) {
      state.dotShape = "square-nogap";
    }
    if (cfg.cardPad == null || cfg.cardPad === "") {
      if (cfg.cardPadH != null && cfg.cardPadH !== "") state.cardPad = Number(cfg.cardPadH);
      else if (cfg.cardPadV != null && cfg.cardPadV !== "") state.cardPad = Number(cfg.cardPadV);
    }
  }

  function normalizeLoadedState() {
    if (state.mode === "revenue" || state.mode === "next-months" || state.mode === "year-end") state.mode = "year";
    if (state.dotShape === "custom" || state.dotShape === "pill") state.dotShape = "circle";
    if (state.bgType === "preset") state.bgType = "match";
    if (state.cardStyle === "glass") state.cardStyle = "flat";
    if (state.fillFocus !== "past" && state.fillFocus !== "future") state.fillFocus = "past";
    const rot = Number(state.bgRotate);
    state.bgRotate = [0, 90, 180, 270].includes(rot) ? rot : 0;
  }

  function loadSavedConfig() {
    // Studio settings live in localStorage only — never rewrite the page URL.
    const saved = loadFromLocalStorage();
    if (saved) applyLoaded(saved);
    normalizeLoadedState();
  }

  function syncForm() {
    const set = (id, val) => {
      const el = document.getElementById(id);
      if (!el || val === undefined || val === null) return;
      if (el.type === "checkbox") el.checked = !!val;
      else el.value = val;
    };
    set("fg", state.fg);
    set("fgOther", state.fgOther);
    set("highlightColor", state.highlightColor);
    set("bg", state.bg);
    set("birthday", state.birthday);
    set("lifespan", state.lifespan);
    set("startDate", state.startDate);
    set("endDate", state.endDate);
    set("customTitle", state.customTitle);
    set("bgUrl", state.bgUrl);
    set("bgBlur", state.bgBlur);
    set("bgDim", state.bgDim);
    set("bgFocalX", state.bgFocalX);
    set("bgFocalY", state.bgFocalY);
    set("gridOpacity", state.gridOpacity);
    set("cardOpacity", state.cardOpacity);
    set("cardRadius", state.cardRadius);
    set("cardPad", state.cardPad);
    set("cardFade", state.cardFade);
    set("margin", state.margin);
    set("marginTop", state.marginTop);
    set("dotGap", state.dotGap);
    set("dotSize", state.dotSize);
    set("showFooter", state.showFooter);
    set("showCard", state.showCard);
    set("weekStartsMonday", state.weekStart !== "sunday");
    set("showAxes", state.showAxes);
    set("showTitle", state.showTitle);
    set("muteText", state.muteText);
    set("titleSize", state.titleSize);
    set("footerSize", state.footerSize);
    const device = $("#device");
    if (device && state.deviceId) {
      device.value = state.deviceId;
      const opt = device.selectedOptions[0];
      if (opt?.dataset.w && opt?.dataset.h) {
        state.width = Number(opt.dataset.w);
        state.height = Number(opt.dataset.h);
      }
    }
    updatePhoneAspect();
    document.querySelectorAll("[data-seg]").forEach((btn) => {
      const key = btn.dataset.seg;
      const map = {
        mode: "mode",
        shape: "dotShape",
        grouping: "grouping",
        bgType: "bgType",
        fillFocus: "fillFocus",
        bgRotate: "bgRotate",
      };
      const stateKey = map[key];
      if (!stateKey) return;
      btn.classList.toggle("active", btn.dataset.value === String(state[stateKey]));
    });
  }

  function boolParam(v) {
    return v ? "1" : "0";
  }

  function flatQuery(cfg) {
    const p = new URLSearchParams();
    const set = (k, v) => {
      if (v === undefined || v === null || v === "") return;
      p.set(k, String(v));
    };
    const mode = cfg.mode || "year";
    set("mode", mode);

    // Calendar inputs that matter for the selected mode only.
    switch (mode) {
      case "life":
        set("birthday", cfg.birthday);
        set("lifespan", cfg.lifespan);
        set("grouping", cfg.grouping);
        set("week_start", cfg.weekStart);
        break;
      case "year":
        set("year", cfg.year || new Date().getFullYear());
        break;
      case "year-months":
        set("year", cfg.year || new Date().getFullYear());
        set("week_start", cfg.weekStart);
        break;
      case "quarters":
        set("year", cfg.year || new Date().getFullYear());
        break;
      case "monthly":
      case "weekly":
        set("week_start", cfg.weekStart);
        break;
      case "custom":
      case "goal":
        set("start_date", cfg.startDate);
        set("end_date", cfg.endDate);
        set("week_start", cfg.weekStart);
        break;
      case "habits":
        if (cfg.habits && cfg.habits.length) {
          set("habits", JSON.stringify(cfg.habits));
        }
        break;
      default:
        break;
    }

    set("title", cfg.customTitle);
    p.set("show_title", boolParam(cfg.showTitle));
    if (cfg.showTitle) {
      set("title_size", cfg.titleSize);
    }
    p.set("show_footer", boolParam(cfg.showFooter));
    if (cfg.showFooter) {
      set("footer_size", cfg.footerSize);
    }
    p.set("mute_text", boolParam(cfg.muteText));

    if (["life", "year", "year-months", "monthly", "weekly", "habits"].includes(mode)) {
      p.set("show_axes", boolParam(cfg.showAxes));
    }

    set("fg", cfg.fg);
    set("fg_other", cfg.fgOther);
    set("highlight_color", cfg.highlightColor);
    set("fill_focus", cfg.fillFocus || "past");
    set("shape", shapeForAPI(cfg.dotShape));
    if (noGapForAPI(cfg.dotShape)) p.set("no_gap", "1");
    else set("dot_gap", cfg.dotGap);
    set("dot_size", cfg.dotSize);
    p.set("highlight_today", boolParam(cfg.highlightToday));

    set("bg_type", cfg.bgType || "match");
    if (cfg.bgType === "solid") {
      set("bg", cfg.bg);
    } else if (cfg.bgType === "url") {
      set("bg_url", cfg.bgUrl);
      set("bg_rotate", cfg.bgRotate || 0);
      set("bg_blur", cfg.bgBlur);
      set("bg_dim", cfg.bgDim);
      set("bg_focal_x", cfg.bgFocalX);
      set("bg_focal_y", cfg.bgFocalY);
    }

    set("grid_opacity", cfg.gridOpacity);
    p.set("show_card", boolParam(cfg.showCard));
    if (cfg.showCard) {
      set("card_opacity", cfg.cardOpacity);
      set("card_radius", cfg.cardRadius);
      set("card_pad_h", cfg.cardPad);
      set("card_pad_v", cfg.cardPad);
      set("card_fade", cfg.cardFade);
    }

    set("device", cfg.deviceId);
    set("width", cfg.width);
    set("height", cfg.height);
    set("screen", "lock");
    set("margin", cfg.margin);
    set("margin_top", cfg.marginTop);
    return p;
  }

  function updatePhoneAspect() {
    const phone = document.querySelector(".phone");
    if (!phone || !state.width || !state.height) return;
    phone.style.setProperty("--phone-ar", `${state.width} / ${state.height}`);
  }

  function studioQuery() {
    return flatQuery(state).toString();
  }

  function wallpaperURL(opts = {}) {
    let cfg = state;
    if (opts.preview) {
      const maxEdge = 640;
      const scale = Math.min(1, maxEdge / Math.max(state.width, state.height));
      cfg = {
        ...state,
        width: Math.max(120, Math.round(state.width * scale)),
        height: Math.max(240, Math.round(state.height * scale)),
      };
    }
    return `/wallpaper.png?${flatQuery(cfg)}`;
  }

  function refreshPreview() {
    saveToLocalStorage();
    updatePhoneAspect();
    const u = wallpaperURL({ preview: true }) + "&_=" + Date.now();
    preview.src = u;
    loupeImg.src = u;
    const abs = location.origin + wallpaperURL();
    urlBox.textContent = abs;
    qrImg.src = "/api/qr?url=" + encodeURIComponent(abs);
    updateModePanels();
  }

  function updateModePanels() {
    document.querySelectorAll("[data-mode-panel]").forEach((el) => {
      const modes = el.getAttribute("data-mode-panel").split(",").map((m) => m.trim());
      el.hidden = !modes.includes(state.mode);
    });
    const bgUrlFields = document.getElementById("bgUrlFields");
    if (bgUrlFields) bgUrlFields.hidden = state.bgType !== "url";
    const bgColorField = document.getElementById("bgColorField");
    if (bgColorField) bgColorField.hidden = state.bgType !== "solid";
    const cardFields = document.getElementById("cardFields");
    if (cardFields) cardFields.hidden = !state.showCard;
    const titleFields = document.getElementById("titleFields");
    if (titleFields) titleFields.hidden = !state.showTitle;
    const footerSizeField = document.getElementById("footerSizeField");
    if (footerSizeField) footerSizeField.hidden = !state.showFooter;
  }

  const lockDate = document.getElementById("lockDate");
  if (lockDate) {
    lockDate.textContent = new Date().toLocaleDateString(undefined, {
      weekday: "short",
      month: "short",
      day: "numeric",
    });
  }

  function showToast(msg) {
    toast.textContent = msg;
    toast.classList.add("show");
    setTimeout(() => toast.classList.remove("show"), 1800);
  }

  function bindSeg(name, key) {
    document.querySelectorAll(`[data-seg="${name}"]`).forEach((btn) => {
      btn.addEventListener("click", () => {
        document.querySelectorAll(`[data-seg="${name}"]`).forEach((b) => b.classList.remove("active"));
        btn.classList.add("active");
        state[key] = btn.dataset.value;
        refreshPreview();
      });
    });
  }

  bindSeg("mode", "mode");
  bindSeg("shape", "dotShape");
  bindSeg("grouping", "grouping");
  bindSeg("bgType", "bgType");
  bindSeg("fillFocus", "fillFocus");
  document.querySelectorAll('[data-seg="bgRotate"]').forEach((btn) => {
    btn.addEventListener("click", () => {
      document.querySelectorAll('[data-seg="bgRotate"]').forEach((b) => b.classList.remove("active"));
      btn.classList.add("active");
      state.bgRotate = Number(btn.dataset.value) || 0;
      refreshPreview();
    });
  });

  const fields = [
    ["birthday", "birthday"],
    ["lifespan", "lifespan", Number],
    ["startDate", "startDate"],
    ["endDate", "endDate"],
    ["customTitle", "customTitle"],
    ["bgUrl", "bgUrl"],
    ["bgBlur", "bgBlur", Number],
    ["bgDim", "bgDim", Number],
    ["bgFocalX", "bgFocalX", Number],
    ["bgFocalY", "bgFocalY", Number],
    ["gridOpacity", "gridOpacity", Number],
    ["cardOpacity", "cardOpacity", Number],
    ["cardRadius", "cardRadius", Number],
    ["cardPad", "cardPad", Number],
    ["cardFade", "cardFade", Number],
    ["margin", "margin", Number],
    ["marginTop", "marginTop", Number],
    ["dotGap", "dotGap", Number],
    ["dotSize", "dotSize", Number],
    ["titleSize", "titleSize", Number],
    ["footerSize", "footerSize", Number],
    ["fg", "fg"],
    ["fgOther", "fgOther"],
    ["bg", "bg"],
    ["highlightColor", "highlightColor"],
  ];
  fields.forEach(([id, key, cast]) => {
    const el = document.getElementById(id);
    if (!el) return;
    el.addEventListener("change", () => {
      state[key] = cast ? cast(el.value) : el.value;
      refreshPreview();
    });
    el.addEventListener("input", () => {
      if (el.type === "range" || el.type === "number" || el.type === "color") {
        state[key] = cast ? cast(el.value) : el.value;
        refreshPreview();
      }
    });
  });

  $("#device")?.addEventListener("change", (e) => {
    const opt = e.target.selectedOptions[0];
    state.deviceId = e.target.value;
    state.width = Number(opt.dataset.w);
    state.height = Number(opt.dataset.h);
    updatePhoneAspect();
    refreshPreview();
  });

  $("#showCard")?.addEventListener("change", (e) => {
    state.showCard = e.target.checked;
    refreshPreview();
  });
  $("#weekStartsMonday")?.addEventListener("change", (e) => {
    state.weekStart = e.target.checked ? "monday" : "sunday";
    refreshPreview();
  });
  $("#showAxes")?.addEventListener("change", (e) => {
    state.showAxes = e.target.checked;
    state.showLabels = e.target.checked;
    refreshPreview();
  });
  $("#showTitle")?.addEventListener("change", (e) => {
    state.showTitle = e.target.checked;
    refreshPreview();
  });
  $("#showFooter")?.addEventListener("change", (e) => {
    state.showFooter = e.target.checked;
    refreshPreview();
  });
  $("#muteText")?.addEventListener("change", (e) => {
    state.muteText = e.target.checked;
    refreshPreview();
  });
  $("#btnCopy")?.addEventListener("click", async () => {
    const btn = $("#btnCopy");
    const label = btn?.dataset.label || btn?.textContent || "Copy wallpaper URL";
    if (btn && !btn.dataset.label) btn.dataset.label = label;
    let ok = false;
    try {
      await navigator.clipboard.writeText(urlBox.textContent);
      ok = true;
    } catch {
      /* clipboard unavailable */
    }
    if (!btn) return;
    btn.textContent = ok ? "Copied!" : "Copy failed";
    btn.classList.toggle("copied", ok);
    clearTimeout(btn._copyTimer);
    btn._copyTimer = setTimeout(() => {
      btn.textContent = btn.dataset.label;
      btn.classList.remove("copied");
    }, 1600);
  });

  const resetModal = document.getElementById("resetModal");

  function openResetModal() {
    if (!resetModal) return;
    resetModal.hidden = false;
    document.getElementById("btnResetConfirm")?.focus();
  }

  function closeResetModal() {
    if (!resetModal) return;
    resetModal.hidden = true;
  }

  function applyResetDefaults() {
    try {
      localStorage.removeItem(STORAGE_KEY);
    } catch {
      /* ignore */
    }
    const next = structuredClone(defaults);
    next.habits = (next.habits || []).map((h) => ({
      ...h,
      completions: {},
    }));
    Object.assign(state, next);
    syncForm();
    refreshPreview();
    showToast("Reset to defaults");
  }

  function parseBoolParam(v) {
    return v === "1" || v === "true";
  }

  function configFromQuery(params) {
    const get = (k) => {
      const v = params.get(k);
      return v == null ? "" : v;
    };
    const num = (k) => {
      const v = get(k);
      if (v === "") return undefined;
      const n = Number(v);
      return Number.isFinite(n) ? n : undefined;
    };
    const bool = (k) => {
      const v = get(k);
      if (v === "") return undefined;
      return parseBoolParam(v);
    };
    const str = (k) => {
      const v = get(k);
      return v === "" ? undefined : v;
    };

    const cfg = {};
    const set = (key, val) => {
      if (val !== undefined) cfg[key] = val;
    };

    set("mode", str("mode"));
    set("birthday", str("birthday") || str("dob"));
    set("lifespan", num("lifespan"));
    set("year", num("year"));
    set("startDate", str("start_date"));
    set("endDate", str("end_date") || str("goal_date"));
    set("customTitle", str("title"));
    set("fg", str("fg"));
    set("fgOther", str("fg_other"));
    set("bg", str("bg"));
    set("highlightColor", str("highlight_color"));
    set("grouping", str("grouping"));
    set("weekStart", str("week_start"));
    set("dotShape", str("shape"));
    set("shapeStyle", str("shape_style"));
    set("noGap", bool("no_gap"));
    set("dotGap", num("dot_gap"));
    set("dotSize", num("dot_size"));
    set("fillFocus", str("fill_focus"));
    set("highlightToday", bool("highlight_today"));
    set("showNumbers", bool("show_numbers"));
    set("showBoth", bool("show_both"));
    set("bgType", str("bg_type"));
    set("bgUrl", str("bg_url"));
    set("bgRotate", num("bg_rotate"));
    set("bgBlur", num("bg_blur"));
    set("bgDim", num("bg_dim"));
    set("bgFocalX", num("bg_focal_x"));
    set("bgFocalY", num("bg_focal_y"));
    set("gridOpacity", num("grid_opacity"));
    set("showCard", bool("show_card"));
    set("cardStyle", str("card_style"));
    set("cardOpacity", num("card_opacity"));
    set("cardRadius", num("card_radius"));
    set("cardFade", num("card_fade"));
    set("cardPadH", num("card_pad_h"));
    set("cardPadV", num("card_pad_v"));
    set("showAxes", bool("show_axes"));
    set("showTitle", bool("show_title"));
    set("showLabels", bool("show_labels"));
    set("showBranding", bool("show_branding"));
    set("showProgress", bool("show_progress"));
    set("muteText", bool("mute_text"));
    if (params.has("show_footer")) {
      set("showFooter", bool("show_footer"));
    } else if (params.has("footer_mode")) {
      // Legacy wallpaper URLs used footer_mode=none|numbers.
      set("showFooter", get("footer_mode") !== "none");
    }
    set("titleSize", num("title_size"));
    set("footerSize", num("footer_size"));
    set("deviceId", str("device"));
    set("width", num("width"));
    set("height", num("height"));
    set("widgetLayout", str("widget_layout"));
    set("margin", num("margin"));
    set("marginTop", num("margin_top"));
    set("statsScale", num("stats_scale"));

    const habitsRaw = get("habits");
    if (habitsRaw) {
      try {
        cfg.habits = JSON.parse(habitsRaw);
      } catch {
        /* ignore bad habits JSON */
      }
    }

    if (cfg.showAxes !== undefined && cfg.showLabels === undefined) {
      cfg.showLabels = cfg.showAxes;
    }
    return cfg;
  }

  function parseImportURL(raw) {
    const text = String(raw || "").trim();
    if (!text) throw new Error("Paste a wallpaper URL first.");
    let url;
    try {
      url = new URL(text, location.origin);
    } catch {
      throw new Error("That doesn’t look like a valid URL.");
    }
    const params = url.searchParams;
    if (![...params.keys()].length) {
      throw new Error("URL has no config query parameters.");
    }
    return configFromQuery(params);
  }

  function applyImportFromURL(raw) {
    const cfg = parseImportURL(raw);
    try {
      localStorage.clear();
    } catch {
      /* ignore */
    }
    const next = structuredClone(defaults);
    next.habits = (next.habits || []).map((h) => ({
      ...h,
      completions: {},
    }));
    Object.assign(state, next);
    applyLoaded(cfg);
    normalizeLoadedState();
    syncForm();
    refreshPreview();
    showToast("Imported from URL");
  }

  const importModal = document.getElementById("importModal");
  const importUrlInput = document.getElementById("importUrl");
  const importError = document.getElementById("importError");

  function openImportModal() {
    if (!importModal) return;
    if (importError) {
      importError.hidden = true;
      importError.textContent = "";
    }
    if (importUrlInput) importUrlInput.value = "";
    importModal.hidden = false;
    importUrlInput?.focus();
  }

  function closeImportModal() {
    if (!importModal) return;
    importModal.hidden = true;
  }

  $("#btnReset")?.addEventListener("click", openResetModal);
  $("#btnResetConfirm")?.addEventListener("click", () => {
    closeResetModal();
    applyResetDefaults();
  });
  resetModal?.querySelectorAll("[data-close-reset]").forEach((el) => {
    el.addEventListener("click", closeResetModal);
  });

  $("#btnImport")?.addEventListener("click", (e) => {
    e.preventDefault();
    openImportModal();
  });
  $("#btnImportConfirm")?.addEventListener("click", () => {
    try {
      applyImportFromURL(importUrlInput?.value);
      closeImportModal();
    } catch (err) {
      if (importError) {
        importError.hidden = false;
        importError.textContent = err.message || "Import failed.";
      }
    }
  });
  importModal?.querySelectorAll("[data-close-import]").forEach((el) => {
    el.addEventListener("click", closeImportModal);
  });

  document.addEventListener("keydown", (e) => {
    if (e.key !== "Escape") return;
    if (resetModal && !resetModal.hidden) closeResetModal();
    if (importModal && !importModal.hidden) closeImportModal();
  });

  const screen = $(".phone-screen");
  screen?.addEventListener("mousemove", (e) => {
    const rect = screen.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    loupe.style.display = "block";
    loupe.style.left = x - 45 + "px";
    loupe.style.top = y - 45 + "px";
    loupeImg.style.width = rect.width * 2 + "px";
    loupeImg.style.height = rect.height * 2 + "px";
    loupeImg.style.left = -x * 2 + 45 + "px";
    loupeImg.style.top = -y * 2 + 45 + "px";
  });
  screen?.addEventListener("mouseleave", () => {
    loupe.style.display = "none";
  });

  $("#btnLogHabit")?.addEventListener("click", async () => {
    const today = new Date().toISOString().slice(0, 10);
    const res = await fetch("/api/habits/log?habit=h1&done=1&" + studioQuery(), { method: "POST" });
    if (res.ok) {
      const data = await res.json();
      if (data.config) applyLoaded(data.config);
      else if (state.habits[0]) {
        if (!state.habits[0].completions) state.habits[0].completions = {};
        state.habits[0].completions[today] = true;
      }
      showToast("Habit logged");
      refreshPreview();
    } else {
      showToast("Habit log failed");
    }
  });

  loadSavedConfig();
  syncForm();
  refreshPreview();
})();
