document.getElementById("loginForm").onsubmit = async (e) => {
  e.preventDefault();

  const btn = document.querySelector(".btn-login");
  const errorMsg = document.getElementById("errorMsg");
  const emailVal = document.getElementById("email").value;
  const passwordVal = document.getElementById("password").value;

  // Loading state
  btn.classList.add("loading");
  btn.disabled = true;
  errorMsg.classList.remove("visible");

  try {
    const res = await fetch("/api/v1/users/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: emailVal, password: passwordVal }),
    });

    const data = await res.json();

    if (data.accessToken) {
      document.cookie = "token=" + data.accessToken + "; path=/; SameSite=Strict";
      window.location.href = "/";
    } else {
      errorMsg.textContent = data.message || "Invalid email or password.";
      errorMsg.classList.add("visible");
    }
  } catch (err) {
    errorMsg.textContent = "Something went wrong. Please try again.";
    errorMsg.classList.add("visible");
  } finally {
    btn.classList.remove("loading");
    btn.disabled = false;
  }
};
