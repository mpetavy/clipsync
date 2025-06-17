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
