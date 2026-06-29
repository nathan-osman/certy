(function() {

  /**
   * Handle the dynamic theme switcher and preferences
   */

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

  /**
   * Add copy buttons to text on the page that can be copied
   */

  document.addEventListener("DOMContentLoaded", () => {
    Array.from(document.getElementsByClassName("copyable")).forEach((el) => {
      let b = document.createElement('a')
      b.setAttribute('href', 'javascript:void(0)')
      b.setAttribute('title', "Copy to clipboard")
      let i = document.createElement('i')
      i.classList.add('bi', 'bi-copy')
      b.appendChild(i)
      b.addEventListener('click', () => {
        navigator.clipboard.writeText(el.dataset.value)
      })
      el.appendChild(b)
    })
  })

})()
