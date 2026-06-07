(function() {

  const mql = window.matchMedia("(prefers-color-scheme: dark)")

  let pref = localStorage.getItem("theme") ?? "auto"

  function useDark() {
    switch (pref) {
    case "light":
      return false
    case "dark":
      return true
    default:
      return mql.matches
    }
  }

  function updatePage() {
    if (useDark()) {
      document.documentElement.setAttribute("data-bs-theme", "dark")
    } else {
      document.documentElement.removeAttribute("data-bs-theme")
    }
  }
  updatePage()

  mql.addEventListener("change", updatePage)

  document.addEventListener("DOMContentLoaded", () => {
    for (const el of document.getElementsByClassName("certy-toggle")) {
      el.addEventListener("click", () => {
        pref = el.dataset.certyToggle
        updatePage()
        localStorage.setItem("theme", pref)
      })
    }
  })

})()
