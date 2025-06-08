document.getElementById("set").addEventListener("click", () => {
  chrome.runtime.sendMessage({ action: "set" }, (response) => {
    if (chrome.runtime.lastError) {
      console.error("Error sending message:", chrome.runtime.lastError);
    } else if (response && response.status === "success") {
      console.log("Sync triggered successfully.");
    }
  });
});

document.getElementById("get").addEventListener("click", () => {
  chrome.runtime.sendMessage({ action: "get" }, (response) => {
    if (chrome.runtime.lastError) {
      console.error("Error sending message:", chrome.runtime.lastError);
    } else if (response && response.status === "success") {
      console.log("Sync triggered successfully.");
    }
  });
});
