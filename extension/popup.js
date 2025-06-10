document.getElementById("sync").addEventListener("click", () => {
  chrome.runtime.sendMessage({ action: "sync" }, (response) => {
    if (chrome.runtime.lastError) {
      console.error("Error sending message:", chrome.runtime.lastError);
    } else if (response && response.status === "success") {
      console.log("Sync triggered successfully.");
    }
  });
});
