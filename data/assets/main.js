

//
//   const isSystemDark = window.matchMedia("(prefers-color-scheme: dark)").matches
function toggleDarkMode() {
    const body = document.querySelector("body");
    body.classList.toggle("dark");
    document.cookie = `dark-mode=${body.classList.contains("dark") ? "true" : "false"}; Path=/`;
    setHighlightColors()
}



function setHighlightColors() {
    var link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/base16/solarized-light.min.css';

    if (document.body.classList.contains('dark')) {
        link.href = 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/base16/solarized-dark.min.css';
    }
    document.body.appendChild(link);
}