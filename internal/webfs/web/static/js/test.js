console.log('Simple test script loaded');
alert('Simple test alert');
fetch('/health').then(response => {
    console.log('Simple test request successful:', response.status);
}).catch(error => {
    console.error('Simple test request failed:', error);
});