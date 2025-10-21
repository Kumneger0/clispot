
    // Highlight active sidebar link on scroll
    const sections = document.querySelectorAll('section');
    const navLinks = document.querySelectorAll('.sidebar nav a');
    const menuToggle = document.querySelector('.menu-toggle');
    const sidebarNav = document.querySelector('.sidebar-nav');

// Menu toggle (mobile)
if (menuToggle) {
  menuToggle.addEventListener('click', () => {
    sidebarNav.classList.toggle('active');
  });
}

// Active link highlighting on scroll
    window.addEventListener('scroll', () => {
      let current = '';
      sections.forEach(section => {
        const sectionTop = section.offsetTop - 60;
        if (pageYOffset >= sectionTop) {
          current = section.getAttribute('id');
        }
      });
      navLinks.forEach(link => {
        link.classList.remove('active');
        if (link.getAttribute('href') === '#' + current) {
          link.classList.add('active');
        }
      });
    });

    // Close sidebar when a link is clicked on mobile
    navLinks.forEach(link => {
      link.addEventListener('click', () => {
        if (window.innerWidth <= 900) {
          sidebarNav.classList.remove('active');
        }
      });
    });

// No theme toggle: site is always dark

// Copy buttons for code blocks
const pres = document.querySelectorAll('pre');
pres.forEach(pre => {
  const button = document.createElement('button');
  button.className = 'copy-btn';
  button.type = 'button';
  button.textContent = 'Copy';
  button.addEventListener('click', async () => {
    try {
      const codeEl = pre.querySelector('code');
      const text = codeEl ? codeEl.innerText : pre.innerText.replace(/\s*Copy\s*$/, '');
      await navigator.clipboard.writeText(text);
      const original = button.textContent;
      button.textContent = 'Copied!';
      setTimeout(() => (button.textContent = original), 1200);
    } catch (e) {
      console.error('Copy failed', e);
    }
  });
  pre.appendChild(button);
});
