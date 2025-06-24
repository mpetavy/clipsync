function updateFootnote() {
    const manifest = chrome.runtime.getManifest();

    const footnote = document.getElementById('footnote');
    footnote.textContent = `${manifest.name} ${manifest.version}`;
}

function updateTheme() {
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        document.body.classList.add('dark');
        document.body.classList.remove('light');
    } else {
        document.body.classList.add('light');
        document.body.classList.remove('dark');
    }
}

function showError(text) {
    errorBox.textContent = text
    errorBox.style.display = 'block';

    setTimeout(() => {
        errorBox.style.display = 'none';
    }, 2000)
}

function saveData(storageObj) {
    chrome.storage.local.set(storageObj, function () {
        console.log('All data saved');
    });
}

function loadData() {
    return new Promise((resolve) => {
        chrome.storage.local.get(null, function (result) {
            console.log('All stored data:', result);
            resolve(result);
        });
    });
}

function updateSettings() {
    console.log('updateSettings');

    loadData().then(data => {
        console.log('data: ', data);

        const serverUrl = document.getElementById("server-url");
        const serverUsername = document.getElementById("server-username");
        const serverPassword = document.getElementById("server-password");

        if (data.serverUrl) serverUrl.value = data.serverUrl;
        if (data.serverUsername) serverUsername.value = data.serverUsername;
        if (data.serverPassword) serverPassword.value = data.serverPassword;
    }).catch(error => {
        console.error('Error loading data:', error);
    });
}

updateFootnote();
updateTheme();
updateSettings();

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', updateTheme);

document.getElementById("sync").addEventListener("click", () => {
    const serverUrl = document.getElementById("server-url").value;
    const serverUsername = document.getElementById("server-username").value;
    const serverPassword = document.getElementById("server-password").value;

    saveData({"serverUrl": serverUrl, "serverUsername": serverUsername, "serverPassword": serverPassword});

    try {
        new URL(serverUrl);

        fetch(serverUrl + "/sync",{
            method: "HEAD",
        }).then(response => {
            // Handle the fetch response if needed
        }).catch(error => {
            console.error("Fetch error:", error);
        });
    } catch (e) {
        showError("Please enter a valid URL! (" + e.message + ")");
        return;
    }

    if (!serverPassword || serverPassword.length < 8) {
        showError("Please enter a valid password!");
        return;
    }

    chrome.runtime.sendMessage({
        action: "sync",
        serverUrl: serverUrl,
        serverUsername: serverUsername,
        serverPassword: serverPassword
    }, (response) => {
        if (chrome.runtime.lastError) {
            console.error("Error sending message:", chrome.runtime.lastError);
        } else if (response && response.status === "success") {
            console.log("Sync triggered successfully.");
        }
    });
});
