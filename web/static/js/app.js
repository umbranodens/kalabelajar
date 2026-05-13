document.addEventListener("DOMContentLoaded", () => {
  // ── Form submit loading state ──────────────────────────────
  document.querySelectorAll("form").forEach((form) => {
    form.addEventListener("submit", () => {
      const submitButton = form.querySelector("button[type='submit']");
      if (!submitButton) return;
      submitButton.dataset.originalText = submitButton.textContent;
      submitButton.textContent = "Menyimpan...";
      submitButton.disabled = true;
      form.classList.add("is-submitting");
    });
  });

  // ── Sidebar mobile toggle ──────────────────────────────────
  const hamburgerBtn = document.getElementById("sidebar-toggle");
  const drawer = document.getElementById("sidebar-drawer");
  const overlay = document.getElementById("sidebar-overlay");

  function openDrawer() {
    if (!drawer) return;
    drawer.classList.remove("drawer-closed");
    if (overlay) {
      overlay.classList.remove("opacity-0", "pointer-events-none");
      overlay.classList.add("opacity-100");
    }
  }

  function closeDrawer() {
    if (!drawer) return;
    drawer.classList.add("drawer-closed");
    if (overlay) {
      overlay.classList.add("opacity-0", "pointer-events-none");
      overlay.classList.remove("opacity-100");
    }
  }

  if (hamburgerBtn) {
    hamburgerBtn.addEventListener("click", openDrawer);
  }
  if (overlay) {
    overlay.addEventListener("click", closeDrawer);
  }

  // Close drawer on ESC
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") closeDrawer();
  });
});
