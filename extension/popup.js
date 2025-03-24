document.getElementById("sync").addEventListener("click", () => {
  console.log("Message send");
  // chrome.runtime.sendMessage({ action: "sync" });
});

chrome.runtime.onMessage.addListener((message) => {
  console.log("Message received in popup:", message);
});
