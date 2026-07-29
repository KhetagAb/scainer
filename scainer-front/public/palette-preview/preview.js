(function () {
  function applyTheme(dark) {
    document.documentElement.dataset.theme = dark ? "dark" : "light";
    document.querySelectorAll("[data-theme-btn]").forEach(function (btn) {
      var on = btn.dataset.themeBtn === (dark ? "dark" : "light");
      btn.classList.toggle("active", on);
    });
  }

  document.querySelectorAll("[data-theme-btn]").forEach(function (btn) {
    btn.addEventListener("click", function () {
      applyTheme(btn.dataset.themeBtn === "dark");
    });
  });

  var params = new URLSearchParams(location.search);
  if (params.get("theme") === "dark") applyTheme(true);
  else applyTheme(false);
})();
