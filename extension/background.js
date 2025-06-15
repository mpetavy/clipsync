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

// ##################################################################################################################

async function setBookmarks() {
    try {
        const rootTree = await chrome.bookmarks.getTree();
        // let otherBookmarksFolder = rootTree[0].children.find(child =>
        //     child.title === "Bookmarks"
        // );

        const jsonString = JSON.stringify(rootTree, null, 2);

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

async function clearAllBookmarks() {
    // Get the root of the bookmark tree
    const rootTree = await chrome.bookmarks.getTree();
    const root = rootTree[0];
    // Remove all children of the root (Bookmarks Bar, Other Bookmarks, etc.)
    for (const child of root.children) {
        if (child.title != "" && child.title != "Bookmarks" && child.title != "Other bookmarks") {
            await chrome.bookmarks.removeTree(child.id);
        }
    }
}

async function getBookmarks() {
    try {
        // 1. Remove all current bookmarks except root
        await clearAllBookmarks();

        // 2. Fetch new bookmarks from the server
        const response = await fetch("http://localhost:8080/get");
        let bookmarks = await response.json();
        // Ensure bookmarks is an array (in case server returns a single object)
        if (!Array.isArray(bookmarks)) {
            bookmarks = [bookmarks];
        }

        // 3. Import new bookmarks under "Bookmarks Bar"
        const rootTree = await chrome.bookmarks.getTree();
        const root = rootTree[0];
        let bookmarksBar = rootTree[0]
        // Find "Bookmarks Bar" or create it if missing
        // let bookmarksBar = root.children.find(child => child.title === "Bookmarks");
        // if (!bookmarksBar) {
        //     bookmarksBar = await chrome.bookmarks.create({
        //         title: "Bookmarks Bar"
        //     });
        // }
        // Add all new bookmarks under "Bookmarks Bar"
        for (const bookmark of bookmarks) {
            await addBookmark(bookmark, bookmarksBar.id);
        }
    } catch (error) {
        console.error("Error setting bookmarks:", error);
    }
}

// Helper function to add a bookmark or folder (recursively)
async function addBookmark1(bookmark, parentId) {
    const {children, id, folderType, syncing, dateGroupModified, dateAdded, ...createDetails} = bookmark;
    // const { children, id, ...createDetails } = bookmark;
    if (parentId) createDetails.parentId = parentId;
    const result = await chrome.bookmarks.create(createDetails);
    if (children && children.length > 0 && !createDetails.url) {
        for (const child of children) {
            await addBookmark(child, result.id);
        }
    }
}

async function addBookmark(bookmark, parentId) {
    const {children, id, folderType, syncing, dateGroupModified, dateAdded, ...createDetails} = bookmark;
    if (parentId) createDetails.parentId = parentId;

    // Skip if title is empty or not defined
    if (!createDetails.title) {
        console.log('Skipping: title is empty');
        return;
    }

    // For folders (no URL)
    if (!createDetails.url) {
        // Check if folder with same title already exists in this parent
        const existing = await chrome.bookmarks.getChildren(parentId || createDetails.parentId);
        const exists = existing.some(child =>
            child.title === createDetails.title && !child.url
        );
        if (!exists) {
            const result = await chrome.bookmarks.create(createDetails);
            if (children && children.length > 0) {
                for (const child of children) {
                    await addBookmark(child, result.id);
                }
            }
        } else {
            console.log('Skipping: folder already exists');
        }
        return;
    }

    // For bookmarks (with URL), check if already exists in the parent
    const existing = await chrome.bookmarks.search({
        url: createDetails.url,
        title: createDetails.title,
    });
    // Filter to only those in the correct parent folder
    const existsInParent = existing.some(b =>
        b.parentId === (parentId || createDetails.parentId)
    );
    if (!existsInParent) {
        await chrome.bookmarks.create(createDetails);
    } else {
        console.log('Skipping: bookmark already exists');
    }
}
