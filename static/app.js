document.addEventListener('DOMContentLoaded', () => {
    // Tab Switching Logic
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabPanes = document.querySelectorAll('.tab-pane');

    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            tabBtns.forEach(b => b.classList.remove('active'));
            tabPanes.forEach(p => p.classList.remove('active'));
            
            btn.classList.add('active');
            document.getElementById(btn.dataset.tab).classList.add('active');
        });
    });

    // Toast Notification System
    function showToast(message, type = 'success') {
        const container = document.getElementById('toast-container');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        
        const icon = type === 'success' 
            ? '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>'
            : '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>';
        
        toast.innerHTML = `${icon} <span>${message}</span>`;
        container.appendChild(toast);
        
        setTimeout(() => {
            toast.remove();
        }, 3000);
    }

    // --- Text Area Logic ---
    const textarea = document.getElementById('shared-textarea');
    const saveTextBtn = document.getElementById('save-text-btn');
    const textStatus = document.getElementById('text-status');
    let typingTimer;

    // Load initial text
    fetch('/api/text')
        .then(res => res.text())
        .then(text => {
            textarea.value = text;
        })
        .catch(err => showToast('Failed to load text', 'error'));

    function saveText() {
        textStatus.textContent = 'Saving...';
        textStatus.classList.add('saving');
        
        fetch('/api/text', {
            method: 'POST',
            body: textarea.value
        })
        .then(res => {
            if (res.ok) {
                textStatus.textContent = 'Synced';
                textStatus.classList.remove('saving');
            } else {
                throw new Error('Failed');
            }
        })
        .catch(err => {
            textStatus.textContent = 'Error saving';
            textStatus.classList.remove('saving');
            showToast('Failed to save text', 'error');
        });
    }

    saveTextBtn.addEventListener('click', saveText);

    // Auto-save when typing stops
    textarea.addEventListener('input', () => {
        textStatus.textContent = 'Unsaved changes';
        textStatus.classList.add('saving');
        clearTimeout(typingTimer);
        typingTimer = setTimeout(saveText, 1000);
    });

    // --- File Storage Logic ---
    const dropZone = document.getElementById('drop-zone');
    const fileInput = document.getElementById('file-input');
    const fileList = document.getElementById('file-list');
    const refreshBtn = document.getElementById('refresh-files-btn');

    function formatBytes(bytes, decimals = 2) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const dm = decimals < 0 ? 0 : decimals;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
    }

    function loadFiles() {
        fileList.innerHTML = '<div class="empty-state">Loading files...</div>';
        fetch('/api/files')
            .then(res => res.json())
            .then(files => {
                fileList.innerHTML = '';
                if (!files || files.length === 0) {
                    fileList.innerHTML = '<div class="empty-state">No files uploaded yet.</div>';
                    return;
                }
                
                files.forEach(file => {
                    const li = document.createElement('li');
                    li.className = 'file-item glass-panel';
                    
                    const date = new Date(file.time).toLocaleString();
                    
                    li.innerHTML = `
                        <a href="/api/download/${file.name}" class="file-info" target="_blank" download>
                            <svg class="file-icon" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path><polyline points="13 2 13 9 20 9"></polyline></svg>
                            <div class="file-details">
                                <span class="file-name" title="${file.name}">${file.name}</span>
                                <span class="file-meta">${formatBytes(file.size)} • ${date}</span>
                            </div>
                        </a>
                        <div class="file-actions">
                            <a href="/api/download/${file.name}" download title="Download">
                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                            </a>
                            <button class="delete-btn" data-filename="${file.name}" title="Delete">
                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                            </button>
                        </div>
                    `;
                    
                    const deleteBtn = li.querySelector('.delete-btn');
                    deleteBtn.addEventListener('click', (e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        deleteFile(file.name);
                    });
                    
                    fileList.appendChild(li);
                });
            })
            .catch(err => {
                fileList.innerHTML = '<div class="empty-state">Error loading files</div>';
                showToast('Failed to load files', 'error');
            });
    }

    function deleteFile(filename) {
        if (!confirm(`Are you sure you want to delete ${filename}?`)) return;
        
        fetch(`/api/files/${filename}`, {
            method: 'DELETE'
        })
        .then(res => {
            if (res.ok) {
                showToast('File deleted successfully');
                loadFiles();
            } else {
                throw new Error('Failed to delete');
            }
        })
        .catch(err => showToast('Error deleting file', 'error'));
    }

    function uploadFiles(files) {
        if (files.length === 0) return;
        
        Array.from(files).forEach(file => {
            const formData = new FormData();
            formData.append('file', file);
            
            showToast(`Uploading ${file.name}...`);
            
            fetch('/api/upload', {
                method: 'POST',
                body: formData
            })
            .then(res => {
                if (res.ok) {
                    showToast(`${file.name} uploaded successfully`);
                    loadFiles();
                } else {
                    throw new Error('Upload failed');
                }
            })
            .catch(err => showToast(`Failed to upload ${file.name}`, 'error'));
        });
    }

    // Drag and drop event listeners
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, preventDefaults, false);
    });

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    ['dragenter', 'dragover'].forEach(eventName => {
        dropZone.addEventListener(eventName, () => dropZone.classList.add('dragover'), false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, () => dropZone.classList.remove('dragover'), false);
    });

    dropZone.addEventListener('drop', (e) => {
        const dt = e.dataTransfer;
        const files = dt.files;
        uploadFiles(files);
    });

    dropZone.addEventListener('click', () => {
        fileInput.click();
    });

    fileInput.addEventListener('change', function() {
        uploadFiles(this.files);
    });

    refreshBtn.addEventListener('click', loadFiles);

    // Initial load
    loadFiles();
});
