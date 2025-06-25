function updateLegend() {
    const manifest = chrome.runtime.getManifest();

    const legend = document.getElementById('legend');
    legend.textContent = `${manifest.name} ${manifest.version}`;
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

function saveSettings(storageObj) {
    console.log('saveSettings');

    chrome.storage.local.set(storageObj, function () {
        console.log('settings saved');
    });
}

function loadSettings() {
    console.log('loadSettings');

    return new Promise((resolve) => {
        chrome.storage.local.get(null, function (result) {
            console.log('settings loaded');
            resolve(result);
        });
    });
}

function updateSettings() {
    console.log('updateSettings');

    loadSettings().then(data => {
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
updateSettings();

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', updateTheme);

document.getElementById("sync").addEventListener("click", async () => {
    const serverUrl = document.getElementById("server-url").value;
    const serverUsername = document.getElementById("server-username").value;
    const serverPassword = document.getElementById("server-password").value;

    saveSettings({
        "serverUrl": serverUrl,
        "serverUsername": serverUsername,
        "serverPassword": serverPassword
    });

    try {
        new URL(serverUrl);

        fetch(serverUrl + "/sync", {
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

    const encUsername = await sha256(serverUsername);
    const encPassword = await sha256(serverPassword);

    chrome.runtime.sendMessage({
        action: "sync",
        serverUrl: serverUrl,
        serverUsername: encUsername,
        serverPassword: encPassword
    }, (response) => {
        if (chrome.runtime.lastError) {
            console.error("Error sending message:", chrome.runtime.lastError);
        } else if (response && response.status === "success") {
            console.log("Sync triggered successfully.");
            window.close();
        }
    });
});
