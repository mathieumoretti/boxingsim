// API base URL - adjust this to match your backend server
const API_BASE_URL = 'http://localhost:8080';

// DOM Elements
const loginBtn = document.getElementById('login-btn');
const registerBtn = document.getElementById('register-btn');
const logoutBtn = document.getElementById('logout-btn');
const authSection = document.getElementById('auth-section');
const dashboardSection = document.getElementById('dashboard-section');
const createBoxerSection = document.getElementById('create-boxer-section');

const loginForm = document.getElementById('login-form-content');
const registerForm = document.getElementById('register-form-content');
const boxerForm = document.getElementById('boxer-form');

const createBoxerBtn = document.getElementById('create-boxer-btn');
const startFightBtn = document.getElementById('start-fight-btn');
const resetFightBtn = document.getElementById('reset-fight-btn');

const messageDiv = document.getElementById('message');

// Current user state
let currentUser = null;
let currentBoxers = [];
let selectedBoxer1 = null;
let selectedBoxer2 = null;

// Show message function
function showMessage(text, isSuccess = true) {
    messageDiv.textContent = text;
    messageDiv.className = isSuccess ? 'success' : 'error';

    setTimeout(() => {
        messageDiv.style.display = 'none';
    }, 3000);
}

// Show section function
function showSection(sectionId) {
    // Hide all sections
    document.querySelectorAll('.section').forEach(section => {
        section.style.display = 'none';
    });

    // Show the requested section
    document.getElementById(sectionId).style.display = 'block';
}

// Authentication functions
loginBtn.addEventListener('click', () => {
    showSection('auth-section');
});

registerBtn.addEventListener('click', () => {
    showSection('auth-section');
});

logoutBtn.addEventListener('click', () => {
    currentUser = null;
    localStorage.removeItem('token');
    showSection('auth-section');
    showMessage('Logged out successfully');
});

// Login form submission
loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    try {
        const response = await fetch(`${API_BASE_URL}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
                password: password
            })
        });

        const data = await response.json();

        if (response.ok) {
            currentUser = data.user;
            localStorage.setItem('token', data.token);
            showSection('dashboard-section');
            loadUserBoxers();
            updateUserInfo();
            showMessage('Login successful');
        } else {
            showMessage(data.error || 'Login failed', false);
        }
    } catch (error) {
        showMessage('Network error: ' + error.message, false);
    }
});

// Register form submission
registerForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const username = document.getElementById('reg-username').value;
    const email = document.getElementById('reg-email').value;
    const password = document.getElementById('reg-password').value;

    try {
        const response = await fetch(`${API_BASE_URL}/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
                email: email,
                password: password
            })
        });

        const data = await response.json();

        if (response.ok) {
            showMessage('Registration successful. Please login.');
            showSection('auth-section');
        } else {
            showMessage(data.error || 'Registration failed', false);
        }
    } catch (error) {
        showMessage('Network error: ' + error.message, false);
    }
});

// Boxer creation
createBoxerBtn.addEventListener('click', () => {
    showSection('create-boxer-section');
});

boxerForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    if (!currentUser) {
        showMessage('Please login first', false);
        return;
    }

    const name = document.getElementById('boxer-name').value;
    const className = document.getElementById('boxer-class').value;

    try {
        const response = await fetch(`${API_BASE_URL}/boxers`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                name: name,
                class: className,
                max_health: 100,
                max_energy: 100,
                strength: 10,
                defense: 10,
                agility: 10,
                skill_points: 0
            })
        });

        const data = await response.json();

        if (response.ok) {
            showMessage('Boxer created successfully');
            showSection('dashboard-section');
            loadUserBoxers();
        } else {
            showMessage(data.error || 'Failed to create boxer', false);
        }
    } catch (error) {
        showMessage('Network error: ' + error.message, false);
    }
});

// Load user's boxers
async function loadUserBoxers() {
    if (!currentUser) return;

    try {
        const response = await fetch(`${API_BASE_URL}/users/${currentUser.id}/boxers`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        });

        const data = await response.json();

        if (response.ok) {
            currentBoxers = data;
            renderBoxersList();
        } else {
            showMessage(data.error || 'Failed to load boxers', false);
        }
    } catch (error) {
        showMessage('Network error: ' + error.message, false);
    }
}

// Render boxers list
function renderBoxersList() {
    const boxersList = document.getElementById('boxers-list');

    if (currentBoxers.length === 0) {
        boxersList.innerHTML = '<p>No boxers created yet.</p>';
        return;
    }

    boxersList.innerHTML = currentBoxers.map(boxer => `
        <div class="boxer-card">
            <div class="boxer-card-info">
                <h4>${boxer.name}</h4>
                <div class="boxer-stats">
                    <p>Level: ${boxer.level}</p>
                    <p>Health: ${boxer.health}/${boxer.max_health}</p>
                    <p>Strength: ${boxer.strength}</p>
                </div>
            </div>
        </div>
    `).join('');
}

// Update user info in dashboard
function updateUserInfo() {
    const userInfo = document.getElementById('user-info');
    if (currentUser) {
        userInfo.textContent = `Welcome, ${currentUser.username}!`;
        logoutBtn.style.display = 'inline-block';
        loginBtn.style.display = 'none';
        registerBtn.style.display = 'none';
    } else {
        userInfo.textContent = '';
        logoutBtn.style.display = 'none';
        loginBtn.style.display = 'inline-block';
        registerBtn.style.display = 'inline-block';
    }
}

// Initialize the app
function initApp() {
    // Check if user is already logged in
    const token = localStorage.getItem('token');
    if (token) {
        // Try to validate token by fetching user info
        fetch(`${API_BASE_URL}/users/1`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        }).then(response => {
            if (response.ok) {
                showSection('dashboard-section');
                updateUserInfo();
                loadUserBoxers();
            } else {
                localStorage.removeItem('token');
                showSection('auth-section');
            }
        }).catch(() => {
            localStorage.removeItem('token');
            showSection('auth-section');
        });
    } else {
        showSection('auth-section');
    }
}

// Start the app
initApp();

// Additional fight functionality would go here
startFightBtn.addEventListener('click', async () => {
    if (!selectedBoxer1 || !selectedBoxer2) {
        showMessage('Please select two boxers for the fight', false);
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/boxers/fight`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                boxer1_id: selectedBoxer1.id,
                boxer2_id: selectedBoxer2.id,
                location: "arena"
            })
        });

        const data = await response.json();

        if (response.ok) {
            showMessage('Fight started!');
            // In a real app, you'd display fight results
        } else {
            showMessage(data.error || 'Failed to start fight', false);
        }
    } catch (error) {
        showMessage('Network error: ' + error.message, false);
    }
});

// Reset fight functionality
resetFightBtn.addEventListener('click', () => {
    // Reset any fight-related display elements
    document.getElementById('boxer1-health').textContent = '0';
    document.getElementById('boxer1-energy').textContent = '0';
    document.getElementById('boxer1-strength').textContent = '0';
    document.getElementById('boxer2-health').textContent = '0';
    document.getElementById('boxer2-energy').textContent = '0';
    document.getElementById('boxer2-strength').textContent = '0';
});