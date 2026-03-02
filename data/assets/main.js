

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

// ── Scroll Spy + TOC toggle ────────────────────────────────────────────────

document.addEventListener('DOMContentLoaded', function () {
    initTocToggle();
    initScrollSpy();
});

function initTocToggle() {
    const toggle  = document.getElementById('toc-toggle');
    const sidebar = document.getElementById('toc-sidebar');
    const close   = document.getElementById('toc-close');
    if (!sidebar) return;

    // Create overlay
    const overlay = document.createElement('div');
    overlay.className = 'toc-overlay';
    document.body.appendChild(overlay);

    function openSidebar() {
        sidebar.classList.add('toc-open');
        overlay.classList.add('toc-open');
        document.body.style.overflow = 'hidden';
    }
    function closeSidebar() {
        sidebar.classList.remove('toc-open');
        overlay.classList.remove('toc-open');
        document.body.style.overflow = '';
    }

    if (toggle) toggle.addEventListener('click', openSidebar);
    if (close)  close.addEventListener('click', closeSidebar);
    overlay.addEventListener('click', closeSidebar);

    document.addEventListener('keydown', function (e) {
        if (e.key === 'Escape') closeSidebar();
    });

    // Close on link click when drawer is open (mobile)
    sidebar.querySelectorAll('nav a').forEach(function (a) {
        a.addEventListener('click', function () {
            if (sidebar.classList.contains('toc-open')) closeSidebar();
        });
    });
}

function initScrollSpy() {
    const sidebar = document.getElementById('toc-sidebar');
    if (!sidebar) return;

    // Headings are directly inside .content (blackfriday renders flat)
    const headings = document.querySelectorAll('.content h2[id], .content h3[id]');
    if (headings.length === 0) return;

    const tocLinks = sidebar.querySelectorAll('nav a');
    if (tocLinks.length === 0) return;

    function setActive(id) {
        tocLinks.forEach(function (a) {
            a.classList.remove('toc-active');
        });
        const match = sidebar.querySelector('nav a[href="#' + id + '"]');
        if (match) match.classList.add('toc-active');
    }

    // Use IntersectionObserver; activate the last heading that has crossed
    // the top 30% of the viewport.
    const observer = new IntersectionObserver(function (entries) {
        entries.forEach(function (entry) {
            if (entry.isIntersecting) {
                setActive(entry.target.id);
            }
        });
    }, {
        rootMargin: '0px 0px -70% 0px',
        threshold: 0
    });

    headings.forEach(function (h) { observer.observe(h); });

    // Initialise to first heading
    if (headings[0]) setActive(headings[0].id);
}
