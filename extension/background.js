chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.action === "sync") {
        processBookmarks(sendResponse).catch(error => {
            console.error("Bookmark processing failed:", error);
            sendResponse({status: "error", error: error.message});
        });
        return true; // Indicates async response
    }
});

// chrome.bookmarks.onCreated.addListener(() => {
//     processBookmarks().catch(error => {
//         console.error("Bookmark processing failed:", error);
//     });
// });
//
// chrome.bookmarks.onRemoved.addListener(() => {
//     processBookmarks().catch(error => {
//         console.error("Bookmark processing failed:", error);
//     });
// });

let inProcessBookmarks = 0;

async function processBookmarks(sendResponse) {
    if (inProcessBookmarks === 1) {
        return
    }
    try {
        inProcessBookmarks = 1;

        await backupBookmarks();
        await restoreBookmarks();
        console.log("Bookmarks synced!");
        if (sendResponse) {
            sendResponse({status: "success"});
        }
    } catch (error) {
        inProcessBookmarks = 0;

        console.error("Sync failed:", error);
        if (sendResponse) {
            sendResponse({status: "error", error: error.message});
        }
        throw error; // Re-throw for outer catch
    }

    inProcessBookmarks = 0;
}

// ##################################################################################################################

async function backupBookmarks() {
    const rootTree = await chrome.bookmarks.getTree();

    const jsonString = JSON.stringify(rootTree, null, 2);

    const response = await fetch("http://localhost:8080/backupBookmarks", {
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

async function clearBookmarks() {
    const rootTree = await chrome.bookmarks.getTree();
    const root = rootTree[0];

    const deleteChildren = async (folderId) => {
        const children = await chrome.bookmarks.getChildren(folderId);
        for (const child of children) {
            await chrome.bookmarks.removeTree(child.id); // Removes folder and all descendants
        }
    };

    await deleteChildren(root.children[0].id);
    await deleteChildren(root.children[1].id);

    console.log("Existing bookmarks successfully removed.");
}

// ##################################################################################################################

async function restoreBookmarks() {
    await clearBookmarks();

    const response = await fetch("http://localhost:8080/restoreBookmarks");
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    let bookmarks = await response.json();

    for (const bookmark of bookmarks[0].children) {
        await addBookmark(bookmark, 0);
    }

    console.log("bookmarks successfully created.");
}

async function addBookmark(bookmark, parentId) {
    const {children, id, syncing, dateGroupModified, dateAdded, ...createDetails} = bookmark;
    if (parentId) createDetails.parentId = parentId;

    if (!createDetails.title) {
        console.log('Skipping: title is empty');
        return;
    }

    if (!createDetails.url) {
        const childrenOfParent = await chrome.bookmarks.getChildren(parentId || createDetails.parentId);

        const existingFolder = createDetails.folderType ?
            childrenOfParent.find(child =>
                child.folderType === createDetails.folderType && !child.url
            )
            :
            childrenOfParent.find(child =>
                child.title === createDetails.title && !child.url
            );

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
