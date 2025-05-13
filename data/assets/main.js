

//
//   const isSystemDark = window.matchMedia("(prefers-color-scheme: dark)").matches
function toggleDarkMode() {
    const body = document.querySelector("body");
    body.classList.toggle("dark");
    document.cookie = `dark-mode=${body.classList.contains("dark") ? "true" : "false"}; Path=/`;
}

