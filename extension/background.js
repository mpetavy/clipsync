async function exportBookmarksAsJSON() {
  try {
    const bookmarkTree = await chrome.bookmarks.getTree();

    const jsonString = JSON.stringify(bookmarkTree, null, 2);

    console.log(jsonString);

    const response = await fetch("http://localhost:8080/set", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: jsonString
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    console.log("Bookmarks successfully sent to the server.");
    return jsonString;  } catch (error) {
    console.error("Failed to export bookmarks:", error);
  }
}

async function fetchBookmarks() {
  try {
    const response = await fetch("http://localhost:8080/get");
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
    exportBookmarksAsJSON();

    return true;

    fetchBookmarks().then(() => {
      console.log("Bookmarks synced!");
      sendResponse({ status: "success" });
    });
    return true;
  }
});