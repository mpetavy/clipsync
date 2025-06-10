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
async function replaceBookmarks(bookmarks) {
    // Get the root bookmark folder (usually the "Bookmarks Bar" or similar)
    const root = (await chrome.bookmarks.getTree())[0];

    // Remove all children of the root (usually the Bookmarks Bar and Other Bookmarks)
    // for (const child of root.children) {
    //     await chrome.bookmarks.removeTree(child.id);
    // }

    // Add new bookmarks (assuming the input matches chrome.bookmarks.create format)
    for (const bookmark of bookmarks) {
        await chrome.bookmarks.create(bookmark);
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
