// ##################################################################################################################

async function setBookmarks() {
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
        return jsonString;
    } catch (error) {
        console.error("Failed to export bookmarks:", error);
    }
}

// ##################################################################################################################

async function getBookmarks() {
    try {
        const response = await fetch("http://localhost:8080/get");
        const bookmarks = await response.json();
        await replaceBookmarks(bookmarks);
    } catch (error) {
        console.error("Error fetching bookmarks:", error);
    }
}

// ##################################################################################################################

/**
 * Replaces all current browser bookmarks with the provided list.
 * @param {Array} bookmarks - Array of bookmark objects (as returned by chrome.bookmarks.getTree)
 */
async function addBookmark(bookmark, parentId) {
    const { children, id, folderType, syncing, dateGroupModified, dateAdded, ...createDetails } = bookmark;
    if (parentId) createDetails.parentId = parentId;
    const result = await chrome.bookmarks.create(createDetails);
    if (children && children.length > 0 && !createDetails.url) {
        for (const child of children) {
            await addBookmark(child, result.id);
        }
    }
}

async function replaceBookmarks(bookmarks) {
    // Assuming 'bookmarks' is an array of root children (like the Bookmarks Bar and Other Bookmarks)
    for (const bookmark of bookmarks) {
        await addBookmark(bookmark);
    }
}

// ##################################################################################################################

async function updateBookmarks(bookmarks) {
    await chrome.bookmarks.removeTree((await chrome.bookmarks.getTree())[0].children[0].id);
    for (const bookmark of bookmarks) {
        await chrome.bookmarks.create(bookmark);
    }
}

// ##################################################################################################################

// chrome.alarms.create("syncBookmarks", {periodInMinutes: 60});
// chrome.alarms.onAlarm.addListener((alarm) => {
//     if (alarm.name === "syncBookmarks") {
//         getBookmarks();
//     }
// });

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.action === "sync") {
        // Export bookmarks to server
        setBookmarks().then(() => {
            // Fetch and update bookmarks from server
            return getBookmarks();
        }).then(() => {
            console.log("Bookmarks synced!");
            sendResponse({status: "success"});
        }).catch(error => {
            console.error("Sync failed:", error);
            sendResponse({status: "error", error: error.message});
        });
        // Return true to indicate async response
        return true;
    }
});
