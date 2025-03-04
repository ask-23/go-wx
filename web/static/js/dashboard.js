/**
 * Dashboard JavaScript for go-wx
 * Handles the initialization and updating of charts
 */

// Function to convert wind direction in degrees to cardinal direction
function getWindDirection(degrees) {
    const directions = ['N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE', 'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'];
    const index = Math.round(degrees / 22.5) % 16;
    return directions[index];
}

// Function to format timestamps for charts
function formatTimestamp(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

// Function to format dates for chart labels
function formatDate(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleDateString();
}

// Function to get min/max values from data
function getMinMax(data, property) {
    const values = data.map(item => item[property]).filter(val => val !== null && val !== undefined);
    return {
        min: Math.min(...values),
        max: Math.max(...values)
    };
}

// Function to generate gradient for chart background
function createGradient(ctx, color1, color2) {
    const gradient = ctx.createLinearGradient(0, 0, 0, 300);
    gradient.addColorStop(0, color1);
    gradient.addColorStop(1, color2);
    return gradient;
}

// Initialize all charts with historical data
function initCharts(historyData) {
    // Prepare data for charts
    const labels = historyData.map(item => formatTimestamp(item.timestamp));
    
    // Temperature chart
    const tempCtx = document.getElementById('tempChart').getContext('2d');
    const tempGradient = createGradient(tempCtx, 'rgba(255, 99, 132, 0.2)', 'rgba(255, 255, 255, 0)');
    
    new Chart(tempCtx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Temperature (°F)',
                data: historyData.map(item => item.temperature),
                borderColor: 'rgb(255, 99, 132)',
                backgroundColor: tempGradient,
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: false
                }
            }
        }
    });
    
    // Wind chill chart
    const windChillCtx = document.getElementById('windChillChart').getContext('2d');
    const windChillGradient = createGradient(windChillCtx, 'rgba(75, 192, 192, 0.2)', 'rgba(255, 255, 255, 0)');
    
    new Chart(windChillCtx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Wind Chill (°F)',
                data: historyData.map(item => item.wind_chill),
                borderColor: 'rgb(75, 192, 192)',
                backgroundColor: windChillGradient,
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: false
                }
            }
        }
    });
    
    // Barometer chart
    const baroCtx = document.getElementById('barometerChart').getContext('2d');
    const baroGradient = createGradient(baroCtx, 'rgba(54, 162, 235, 0.2)', 'rgba(255, 255, 255, 0)');
    
    new Chart(baroCtx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Pressure (mbar)',
                data: historyData.map(item => item.pressure),
                borderColor: 'rgb(54, 162, 235)',
                backgroundColor: baroGradient,
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: false
                }
            }
        }
    });
    
    // Rain chart
    const rainCtx = document.getElementById('rainChart').getContext('2d');
    const rainGradient = createGradient(rainCtx, 'rgba(153, 102, 255, 0.2)', 'rgba(255, 255, 255, 0)');
    
    new Chart(rainCtx, {
        type: 'bar',
        data: {
            labels: labels,
            datasets: [{
                label: 'Rain (in)',
                data: historyData.map(item => item.rain),
                backgroundColor: rainGradient,
                borderColor: 'rgb(153, 102, 255)',
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
    
    // Wind speed chart
    const windSpeedCtx = document.getElementById('windSpeedChart').getContext('2d');
    const windSpeedGradient = createGradient(windSpeedCtx, 'rgba(255, 159, 64, 0.2)', 'rgba(255, 255, 255, 0)');
    
    new Chart(windSpeedCtx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Wind Speed (mph)',
                data: historyData.map(item => item.wind_speed),
                borderColor: 'rgb(255, 159, 64)',
                backgroundColor: windSpeedGradient,
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
    
    // Wind direction chart
    const windDirCtx = document.getElementById('windDirChart').getContext('2d');
    
    new Chart(windDirCtx, {
        type: 'scatter',
        data: {
            datasets: [{
                data: historyData.map(item => ({
                    x: Math.cos(item.wind_direction * Math.PI / 180) * item.wind_speed,
                    y: Math.sin(item.wind_direction * Math.PI / 180) * item.wind_speed
                })),
                backgroundColor: 'rgb(75, 192, 192)',
                pointRadius: 3
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                x: {
                    ticks: {
                        callback: function(value) {
                            if (value === 0) return 'W';
                            if (value === 5) return 'E';
                            return '';
                        }
                    }
                },
                y: {
                    ticks: {
                        callback: function(value) {
                            if (value === 0) return 'S';
                            if (value === 5) return 'N';
                            return '';
                        }
                    }
                }
            }
        }
    });
    
    // UV index chart
    const uvCtx = document.getElementById('uvChart').getContext('2d');
    const uvGradient = createGradient(uvCtx, 'rgba(255, 206, 86, 0.2)', 'rgba(255, 255, 255, 0)');
    
    new Chart(uvCtx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'UV Index',
                data: historyData.map(item => item.uv_index),
                borderColor: 'rgb(255, 206, 86)',
                backgroundColor: uvGradient,
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
    
    // Humidity chart
    const humidityCtx = document.getElementById('humidityChart').getContext('2d');
    const humidityGradient = createGradient(humidityCtx, 'rgba(75, 192, 192, 0.2)', 'rgba(255, 255, 255, 0)');
    
    new Chart(humidityCtx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Humidity (%)',
                data: historyData.map(item => item.humidity),
                borderColor: 'rgb(75, 192, 192)',
                backgroundColor: humidityGradient,
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100
                }
            }
        }
    });
}

// Set up periodic refresh
function setupRefresh() {
    // Refresh data every 5 minutes
    setInterval(() => {
        fetch('/api/current')
            .then(response => response.json())
            .then(data => {
                updateCurrentValues(data);
            })
            .catch(error => console.error('Error fetching current data:', error));
            
        fetch('/api/history')
            .then(response => response.json())
            .then(data => {
                // Update charts with new data
                updateCharts(data);
            })
            .catch(error => console.error('Error fetching history data:', error));
    }, 5 * 60 * 1000); // 5 minutes
}

// Update current values on the dashboard
function updateCurrentValues(data) {
    // Update all the dashboard panels with current data
    document.querySelector('.panel:nth-child(1) .current-value').textContent = `${data.temperature.toFixed(1)}°F`;
    document.querySelector('.panel:nth-child(2) .current-value').textContent = `${Math.round(data.humidity)}%`;
    document.querySelector('.panel:nth-child(3) .current-value').textContent = `${data.pressure.toFixed(1)} mbar`;
    document.querySelector('.panel:nth-child(4) .current-value').textContent = 
        `${data.wind_speed.toFixed(1)} mph ${getWindDirection(data.wind_direction)}`;
    document.querySelector('.panel:nth-child(5) .current-value').textContent = `${data.wind_chill.toFixed(1)}°F`;
    document.querySelector('.panel:nth-child(6) .current-value').textContent = `${data.heat_index.toFixed(1)}°F`;
    document.querySelector('.panel:nth-child(7) .current-value').textContent = `${data.dew_point.toFixed(1)}°F`;
    document.querySelector('.panel:nth-child(8) .current-value').textContent = `${data.uv_index.toFixed(1)}`;
    document.querySelector('.panel:nth-child(9) .current-value').textContent = `${data.rain.toFixed(2)} in`;
}

// Call setup functions when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    setupRefresh();
}); 