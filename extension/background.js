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
    const rootTree = await chrome.bookmarks.getTree();
    // let otherBookmarksFolder = rootTree[0].children.find(child =>
    //     child.title === "Bookmarks"
    // );

    const jsonString = JSON.stringify(rootTree, null, 2);

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
}

// ##################################################################################################################

async function clearAllBookmarks() {
    const rootTree = await chrome.bookmarks.getTree();
    const root = rootTree[0];

    // Find "Bookmarks Bar" and "Other Bookmarks"
    const bookmarksBar = root.children.find(child => child.title === "Bookmarks");
    const otherBookmarks = root.children.find(child => child.title === "Other bookmarks");

    const deleteChildren = async (folderId) => {
        const children = await chrome.bookmarks.getChildren(folderId);
        for (const child of children) {
            await chrome.bookmarks.removeTree(child.id); // Removes folder and all descendants
        }
    };

    if (bookmarksBar) await deleteChildren(bookmarksBar.id);
    if (otherBookmarks) await deleteChildren(otherBookmarks.id);
}

async function getBookmarks() {
    // 1. Remove all current bookmarks except root
    await clearAllBookmarks();

    // 2. Fetch new bookmarks from the server
    const response = await fetch("http://localhost:8080/get");
    if (!response.ok) {
        throw new Error(`Response status: ${response.status}`);
    }

    let bookmarks = await response.json();

    let rootTree = await chrome.bookmarks.getTree();

    console.log(bookmarks);
    console.log(rootTree);

    for (const bookmark of bookmarks[0].children) {
        await addBookmark(bookmark, 0);
    }
    // addBookmark(bookmarks[0].children,0);
}

async function addBookmark(bookmark, parentId) {
    console.log(bookmark.title)

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
        const childrenOfParent = await chrome.bookmarks.getChildren(parentId || createDetails.parentId);
        const existingFolder = childrenOfParent.find(child =>
            child.title === createDetails.title && !child.url
        );

        // Use the existing folder's ID if found, otherwise create new folder
        const folderId = existingFolder ? existingFolder.id :
            (await chrome.bookmarks.create(createDetails)).id;

        if (children && children.length > 0) {
            for (const child of children) {
                await addBookmark(child, folderId);
            }
        }
    }

    if (createDetails.url) {
        await chrome.bookmarks.create(createDetails);
    }
}
