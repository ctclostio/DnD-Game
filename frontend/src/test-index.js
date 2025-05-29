console.log("Test index.js loaded!");

// Simple test without any complex initialization
document.addEventListener('DOMContentLoaded', () => {
    console.log("DOM loaded");
    
    const body = document.body;
    body.innerHTML = `
        <h1>D&D Game Test</h1>
        <div id="test-results">
            <p>JavaScript is working!</p>
            <p>API Base URL: /api/v1</p>
        </div>
        <button onclick="testAPI()">Test API</button>
    `;
    
    window.testAPI = async function() {
        try {
            const response = await fetch('/api/v1/health');
            const data = await response.json();
            document.getElementById('test-results').innerHTML += `
                <p>API Response: ${JSON.stringify(data)}</p>
            `;
        } catch (error) {
            document.getElementById('test-results').innerHTML += `
                <p>API Error: ${error.message}</p>
            `;
        }
    };
});