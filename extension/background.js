chrome.runtime.onMessage.addListener((message, sender, chromeResponse) => {
    if (message.action === "sync") {
        processBookmarks(message, chromeResponse).catch(error => {
            console.error(error);
            chromeResponse({status: "error", error: error.message});
        });
        return true; // Indicates async response
    }
});

// chrome.bookmarks.onCreated.addListener(() => {
//     processBookmarks().catch(error => {
//         console.error(error);
//     });
// });
//
// chrome.bookmarks.onRemoved.addListener(() => {
//     processBookmarks().catch(error => {
//         console.error(error);
//     });
// });

inProcessBookmarks = 0;

async function processBookmarks(message, sendResponse) {
    if (inProcessBookmarks === 1) {
        return
    }

    try {
        inProcessBookmarks = 1;

        if (!message.pluginInitialized) {
            if (!message.serverHasAlreadyBookmarks) {
                await backupBookmarks(message);
            }
        }

        await restoreBookmarks(message);

        console.log("Bookmarks synced!");

        if (sendResponse) {
            sendResponse({status: "success"});
        }
    } catch (error) {
        inProcessBookmarks = 0;

        console.error(error);
        if (sendResponse) {
            sendResponse({status: "error", error: error.message});
        }
        throw error; // Re-throw for outer catch
    }

    inProcessBookmarks = 0;
}

// ##################################################################################################################

async function backupBookmarks(message) {
    const rootTree = await chrome.bookmarks.getTree();

    const jsonString = JSON.stringify(rootTree, null, 2);

    const credentials = btoa(message.username + ":"+ message.password);
    const response = await fetch(message.url + "/api/v1/sync", {
        method: "PUT",
        headers: {
            "Authorization": "Basic " + credentials,
            "Content-Type": "application/json",
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

// chrome.storage.local.set({ url: "https://example.com", username: "myuser" }, () => {
//     console.log("Data saved.");
// });
//
// chrome.storage.sync.set({ url: "https://example.com", username: "myuser" }, () => {
//     console.log("Data saved and synced.");
// });
//
// chrome.storage.local.get(["url", "username"], (result) => {
//     console.log("Saved URL:", result.url);
//     console.log("Saved username:", result.username);
// });
//
// chrome.storage.local.remove(["url"], () => {
//     console.log("URL removed.");
// });
//
// chrome.storage.local.clear(() => {
//     console.log("All data cleared.");
// });

// ##################################################################################################################

async function restoreBookmarks(message) {
    await clearBookmarks();

    const credentials = btoa(message.username + ":"+ message.password);
    const response = await fetch(message.url + "/api/v1/sync", {
        headers: {
            "Authorization": "Basic " + credentials
        },
    });

    bookmarks = await response.json();

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
