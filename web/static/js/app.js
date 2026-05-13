document.addEventListener("DOMContentLoaded", () => {
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
});
