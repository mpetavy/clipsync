serverHasAlreadyBookmarks = false;
pluginInitialized = false;

elementLegend = document.getElementById('legend');
elementUrl = document.getElementById("clipsync-url");
elementUsername = document.getElementById("clipsync-username");
elementPassword = document.getElementById("clipsync-password");
elementButtonSync = document.getElementById("sync")
elementButtonReset = document.getElementById("reset")

function updateLegend() {
    const manifest = chrome.runtime.getManifest();

    elementLegend.textContent = `${manifest.name} ${manifest.version}`;
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
    console.error(text);

    errorBox.textContent = text
    errorBox.style.display = 'block';

    setTimeout(() => {
        errorBox.style.display = 'none';
    }, 2000)
}

function saveSettings(settings) {
    console.log('saveSettings');

    chrome.storage.local.set(settings, function () {
        console.log('settings saved');
    });
}

function loadSettings() {
    console.log('loadSettings');

    return new Promise((resolve) => {
        chrome.storage.local.get(null, function (settings) {
            console.log('settings loaded');
            resolve(settings);
        });
    });
}

function deleteSettings() {
    console.log('deleteSettings');

    chrome.storage.local.clear(function() {
        console.log('settings deleted');
    });
}

function refreshSettings() {
    console.log('refreshSettings');

    elementUrl.value = "http://pop-os:8443";
    elementUsername.value = "petavy@gmx.net";
    elementPassword.value = "11111111";

    pluginInitialized = false;

    loadSettings().then(data => {
        if (data.url) elementUrl.value = data.url;
        if (data.username) elementUsername.value = data.username;
        if (data.password) elementPassword.value = data.password;

        pluginInitialized = data.pluginInitialized;
    }).catch(error => {
        console.error('Error loading data:', error);
    });

    elementUrl.focus();
}

async function sha256(message) {
    // Encode the message as UTF-8
    const encoder = new TextEncoder();
    const data = encoder.encode(message);
    // Hash the data
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    // Convert to hex string
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    return hashHex;
}

updateLegend();
updateTheme();

refreshSettings();

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', updateTheme);

elementButtonSync.addEventListener("click", async () => {
    const url = elementUrl.value;
    const username = elementUsername.value;
    const password = elementPassword.value;

    saveSettings({
        "url": url,
        "username": username,
        "password": password,
        "pluginInitialized": true
    });

    const encUsername = await sha256(username);
    const encPassword = await sha256(password);

    try {
        new URL(url);

        const response = await fetch(url + "/api/v1/sync", {
            headers: {
                "username": encUsername,
                "password": encPassword,
            },
        });

        serverHasAlreadyBookmarks = response.status === 200;
    } catch (e) {
        showError("Please enter a valid URL! (" + e + ")");
        return;
    }

    if (!password || password.length < 8) {
        showError("Please enter a valid password!");
        return;
    }

    chrome.runtime.sendMessage({
        action: "sync",
        url: url,
        username: encUsername,
        password: encPassword,
        serverHasAlreadyBookmarks: serverHasAlreadyBookmarks,
        pluginInitialized: pluginInitialized
    }, (response) => {
        if (chrome.runtime.lastError) {
            console.error("Error sending message:", chrome.runtime.lastError);
        } else if (response && response.status === "success") {
            console.log("Sync triggered successfully.");

            //FIXME
            // window.close();
        }
    });
});

elementButtonReset.addEventListener("click", async () => {
    deleteSettings();

    refreshSettings();
});