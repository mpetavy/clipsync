async function fetchBookmarks() {
  try {
    const response = await fetch("https://example.com/bookmarks.json");
    const bookmarks = await response.json();
    await updateBookmarks(bookmarks);
  } catch (error) {
    console.error("Error fetching bookmarks:", error);
  }
}

async function updateBookmarks(bookmarks) {
  await chrome.bookmarks.removeTree((await chrome.bookmarks.getTree())[0].children[0].id);
  for (const bookmark of bookmarks) {
    await chrome.bookmarks.create(bookmark);
  }
}

chrome.alarms.create("syncBookmarks", { periodInMinutes: 60 });
chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === "syncBookmarks") {
    fetchBookmarks();
  }
});

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.action === "sync") {
    fetchBookmarks().then(() => {
      console.log("Bookmarks synced!");
      sendResponse({ status: "success" });
    });
    return true;
  }
});
