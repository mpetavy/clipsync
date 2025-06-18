function applyTheme() {
  if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    document.body.classList.add('dark');
    document.body.classList.remove('light');
  } else {
    document.body.classList.add('light');
    document.body.classList.remove('dark');
  }
}

// Initial check
applyTheme();

// Listen for changes
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyTheme);

document.getElementById("sync").addEventListener("click", () => {
  const serverUrl = document.getElementById("server-url").value;
  const serverPassword = document.getElementById("server-password").value;

  chrome.runtime.sendMessage({
    action: "sync",
    serverUrl: serverUrl,
    serverPassword: serverPassword
  }, (response) => {
    if (chrome.runtime.lastError) {
      console.error("Error sending message:", chrome.runtime.lastError);
    } else if (response && response.status === "success") {
      console.log("Sync triggered successfully.");
    }
  });
});
