document.addEventListener('DOMContentLoaded', function() {
    const addFeedForm = document.getElementById('add-feed-form');
    const feedUrlInput = document.getElementById('feed-url');
    const feedList = document.getElementById('feed-list');
    const feedContent = document.getElementById('feed-content');
    const currentFeedTitle = document.getElementById('current-feed-title');

    // Check if we have stored feeds
    function checkStoredFeeds() {
        const feedItems = document.querySelectorAll('.feed-item');
        if (feedItems.length === 0) {
            showToast('Welcome! Add your first RSS feed to get started', 'info');
        } else {
            showToast(`Loaded ${feedItems.length} feed subscriptions`, 'success');
        }
    }

    // Call this when the page loads
    setTimeout(checkStoredFeeds, 500);

    // Add a new feed with timeout
    addFeedForm.addEventListener('submit', function(e) {
        e.preventDefault();
        const url = feedUrlInput.value.trim();
        
        if (!url) {
            showToast('Please enter a valid URL', 'danger');
            return;
        }

        const submitBtn = this.querySelector('button[type="submit"]');
        const originalText = submitBtn.innerHTML;
        submitBtn.disabled = true;
        submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Adding...';
        
        // Set a timeout to prevent the UI from being stuck
        const timeoutId = setTimeout(() => {
            submitBtn.disabled = false;
            submitBtn.innerHTML = originalText;
            showToast('Request is taking too long. The feed might still be added in the background. Please check your feed list in a moment.', 'warning');
        }, 15000); // 15 seconds timeout
        
        axios.post('/feeds', { url: url })
            .then(function(response) {
                clearTimeout(timeoutId);
                feedUrlInput.value = '';
                showToast('Feed subscription added successfully!', 'success');
                
                // Add the new feed to the list
                const feed = response.data;
                addFeedToList(feed);
                
                // Auto-select the newly added feed
                const newFeedItem = document.querySelector(`.feed-item[data-url="${feed.URL}"]`);
                if (newFeedItem) {
                    newFeedItem.click();
                }
            })
            .catch(function(error) {
                clearTimeout(timeoutId);
                const errorMessage = error.response?.data?.error || error.message || 'Unknown error';
                showToast('Error adding feed: ' + errorMessage, 'danger');
                console.error('Feed addition error:', error);
            })
            .finally(function() {
                clearTimeout(timeoutId);
                submitBtn.disabled = false;
                submitBtn.innerHTML = originalText;
            });
    });

    // Load feed when clicked
    document.addEventListener('click', function(e) {
        if (e.target.closest('.feed-item')) {
            const feedItem = e.target.closest('.feed-item');
            const url = feedItem.dataset.url;
            
            // Remove active class from all items
            document.querySelectorAll('.feed-item').forEach(item => {
                item.classList.remove('active');
            });
            
            // Add active class to clicked item
            feedItem.classList.add('active');
            
            // On mobile, collapse the sidebar after selecting a feed
            if (window.innerWidth < 768) {
                const navbarToggler = document.querySelector('.navbar-toggler');
                if (navbarToggler && !navbarToggler.classList.contains('collapsed')) {
                    navbarToggler.click();
                }
            }
            
            loadFeed(url);
        }
        
        // Delete feed when delete button is clicked
        if (e.target.closest('.delete-feed')) {
            const deleteBtn = e.target.closest('.delete-feed');
            const url = deleteBtn.dataset.url;
            deleteFeed(url);
            e.stopPropagation();
        }
    });

    // Process HTML content to ensure visibility in dark mode
    function processContentForDarkMode(html) {
        // Create a temporary container to manipulate the HTML
        const tempDiv = document.createElement('div');
        tempDiv.innerHTML = html;
        
        // Remove any inline styles that might interfere with readability
        const elementsWithStyle = tempDiv.querySelectorAll('[style]');
        elementsWithStyle.forEach(el => {
            // Keep only essential styles and remove color/background styles
            const style = el.getAttribute('style');
            const cleanStyle = style
                .replace(/color\s*:\s*[^;]+;?/gi, '')
                .replace(/background(-color)?\s*:\s*[^;]+;?/gi, '');
            
            if (cleanStyle.trim()) {
                el.setAttribute('style', cleanStyle);
            } else {
                el.removeAttribute('style');
            }
        });
        
        // Process specific elements that might have light background or text colors
        const elementsToProcess = tempDiv.querySelectorAll('p, span, div, h1, h2, h3, h4, h5, h6, li');
        elementsToProcess.forEach(el => {
            // Add dark-mode-content class to ensure proper styling
            el.classList.add('dark-mode-content');
        });
        
        return tempDiv.innerHTML;
    }

    // Load feed content - Updated to use query parameters
    function loadFeed(url) {
        // Update feed title to show it's being refreshed
        const feedTitle = document.querySelector(`.feed-item[data-url="${url}"] .feed-title`);
        if (feedTitle) {
            const originalTitle = feedTitle.innerHTML;
            feedTitle.innerHTML = `<i class="bi bi-arrow-repeat me-2 spin"></i>${feedTitle.textContent}`;
        }
        
        // Show loading indicator
        feedContent.innerHTML = `
            <div class="text-center py-5">
                <div class="spinner-border text-primary" role="status">
                    <span class="visually-hidden">Loading...</span>
                </div>
                <p class="mt-3 text-muted">Fetching latest feed content...</p>
                <small class="text-muted">Content is always freshly loaded to ensure you see the latest updates</small>
            </div>
        `;
        
        // Set a timeout for loading
        const timeoutId = setTimeout(() => {
            feedContent.innerHTML = `
                <div class="text-center text-warning py-5">
                    <i class="bi bi-exclamation-triangle display-4 mb-3"></i>
                    <p>Loading is taking longer than expected. Please be patient or try again later.</p>
                </div>
            `;
        }, 10000); // 10 seconds timeout
        
        axios.get('/feed', { params: { url: url } })
            .then(function(response) {
                clearTimeout(timeoutId);
                const feed = response.data;
                
                // Update displayed title with freshness indicator
                const lastUpdated = new Date().toLocaleTimeString();
                currentFeedTitle.innerHTML = `
                    ${feed.Title} 
                    <small class="ms-2 text-muted">
                        <i class="bi bi-clock-history"></i> 
                        Updated at ${lastUpdated}
                    </small>
                `;
                
                // Restore original feed list item title
                if (feedTitle) {
                    feedTitle.innerHTML = `<i class="bi bi-rss me-2"></i>${feed.Title}`;
                }
                
                let html = '';
                if (feed.Items && feed.Items.length > 0) {
                    html += '<div class="feed-items-container">';
                    feed.Items.forEach(item => {
                        const date = new Date(item.PublishedAt);
                        const formattedDate = date.toLocaleString();
                        
                        // Process the description/content for better visibility in dark mode
                        let processedContent = item.Description || item.Content || '';
                        processedContent = processContentForDarkMode(processedContent);
                        
                        html += `
                            <div class="feed-item-entry">
                                <h5>
                                    <a href="${item.Link}" target="_blank" rel="noopener noreferrer">
                                        ${item.Title}
                                        <i class="bi bi-box-arrow-up-right ms-1 small"></i>
                                    </a>
                                </h5>
                                <div class="feed-item-meta">
                                    <i class="bi bi-calendar me-1"></i> <span>${formattedDate}</span>
                                </div>
                                <div class="feed-item-content">
                                    ${processedContent}
                                </div>
                            </div>
                        `;
                    });
                    html += '</div>';
                } else {
                    html = `
                        <div class="text-center text-muted py-5">
                            <i class="bi bi-info-circle display-4 mb-3"></i>
                            <p>No items found in this feed</p>
                        </div>
                    `;
                }
                
                feedContent.innerHTML = html;
                
                // Adjust image sizes for better mobile display and add dark mode overlay
                setTimeout(() => {
                    const images = feedContent.querySelectorAll('img');
                    images.forEach(img => {
                        img.style.maxWidth = '100%';
                        img.style.height = 'auto';
                        // Add slight filter to brighten dark images in dark mode
                        img.style.filter = 'brightness(1.1) contrast(0.95)';
                    });
                    
                    // Make sure all links are properly styled
                    const links = feedContent.querySelectorAll('a');
                    links.forEach(link => {
                        link.style.color = 'var(--accent-color)';
                        link.style.textDecoration = 'none';
                    });
                }, 100);
            })
            .catch(function(error) {
                clearTimeout(timeoutId);
                
                // Restore original feed list item title
                if (feedTitle) {
                    feedTitle.innerHTML = `<i class="bi bi-rss me-2"></i>${feedTitle.textContent.trim()}`;
                }
                
                const errorMessage = error.response?.data?.error || error.message || 'Unknown error';
                showToast('Error loading feed: ' + errorMessage, 'danger');
                console.error('Feed loading error:', error);
                feedContent.innerHTML = `
                    <div class="text-center text-danger py-5">
                        <i class="bi bi-exclamation-triangle display-4 mb-3"></i>
                        <p>Error loading feed: ${errorMessage}</p>
                    </div>
                `;
            });
    }

    // Delete a feed - Updated to use query parameters
    function deleteFeed(url) {
        if (confirm('Are you sure you want to remove this feed subscription?')) {
            const feedItem = document.querySelector(`.feed-item[data-url="${url}"]`);
            if (feedItem) {
                // Add "deleting" visual feedback
                feedItem.classList.add('text-muted');
                const deleteBtn = feedItem.querySelector('.delete-feed');
                if (deleteBtn) {
                    const originalContent = deleteBtn.innerHTML;
                    deleteBtn.disabled = true;
                    deleteBtn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>';
                    
                    // Set a timeout
                    const timeoutId = setTimeout(() => {
                        feedItem.classList.remove('text-muted');
                        deleteBtn.disabled = false;
                        deleteBtn.innerHTML = originalContent;
                        showToast('Delete request is taking too long. The server might still be processing it.', 'warning');
                    }, 10000);
                    
                    axios.delete('/feed', { params: { url: url } })
                        .then(function() {
                            clearTimeout(timeoutId);
                            // Remove the feed from the DOM
                            if (feedItem) {
                                // Add animation before removing
                                feedItem.style.transition = 'all 0.3s ease';
                                feedItem.style.opacity = '0';
                                feedItem.style.height = '0';
                                
                                setTimeout(() => {
                                    feedItem.remove();
                                }, 300);
                            }
                            
                            // Clear feed content if it was the active feed
                            if (feedItem && feedItem.classList.contains('active')) {
                                currentFeedTitle.textContent = 'Select a Feed';
                                feedContent.innerHTML = `
                                    <div class="text-center text-muted py-5">
                                        <i class="bi bi-arrow-left-circle display-4 mb-3"></i>
                                        <p>Select a feed from the list to view its contents</p>
                                    </div>
                                `;
                            }
                            
                            showToast('Feed subscription removed successfully!', 'success');
                        })
                        .catch(function(error) {
                            clearTimeout(timeoutId);
                            // Restore the item styling
                            feedItem.classList.remove('text-muted');
                            deleteBtn.disabled = false;
                            deleteBtn.innerHTML = originalContent;
                            
                            const errorMessage = error.response?.data?.error || error.message || 'Unknown error';
                            showToast('Error removing feed: ' + errorMessage, 'danger');
                            console.error('Feed deletion error:', error);
                        });
                }
            }
        }
    }

    // Add a feed to the list - Updated to use query parameters
    function addFeedToList(feed) {
        // Check if feed already exists in the list, if so, don't add it again
        const existingItem = document.querySelector(`.feed-item[data-url="${feed.URL}"]`);
        if (existingItem) {
            existingItem.classList.add('highlight');
            setTimeout(() => {
                existingItem.classList.remove('highlight');
            }, 2000);
            return;
        }
        
        const li = document.createElement('li');
        li.className = 'list-group-item d-flex justify-content-between align-items-center feed-item';
        li.dataset.url = feed.URL;
        li.innerHTML = `
            <span class="feed-title"><i class="bi bi-rss me-2"></i>${feed.Title}</span>
            <div class="feed-actions">
                <a href="/export?url=${encodeURIComponent(feed.URL)}" class="btn btn-sm btn-outline-secondary" target="_blank" title="View RSS"><i class="bi bi-box-arrow-up-right"></i></a>
                <button class="btn btn-sm btn-danger delete-feed" data-url="${feed.URL}" title="Delete Feed"><i class="bi bi-trash"></i></button>
            </div>
        `;
        
        // Add with animation
        li.style.opacity = '0';
        feedList.appendChild(li);
        
        // Trigger animation
        setTimeout(() => {
            li.style.transition = 'opacity 0.3s ease';
            li.style.opacity = '1';
        }, 10);
    }

    // Show toast notification
    function showToast(message, type) {
        const toast = document.createElement('div');
        toast.className = `toast bg-${type} text-white`;
        toast.setAttribute('role', 'alert');
        toast.setAttribute('aria-live', 'assertive');
        toast.setAttribute('aria-atomic', 'true');
        toast.innerHTML = `
            <div class="toast-header bg-${type} text-white">
                <i class="bi bi-${type === 'success' ? 'check-circle' : type === 'danger' ? 'exclamation-triangle' : type === 'warning' ? 'exclamation-circle' : 'info-circle'} me-2"></i>
                <strong class="me-auto">${type === 'success' ? 'Success' : type === 'danger' ? 'Error' : type === 'warning' ? 'Warning' : 'Information'}</strong>
                <button type="button" class="btn-close btn-close-white" data-bs-dismiss="toast" aria-label="Close"></button>
            </div>
            <div class="toast-body">
                ${message}
            </div>
        `;
        
        document.body.appendChild(toast);
        
        const bsToast = new bootstrap.Toast(toast, {
            autohide: true,
            delay: 4000
        });
        
        bsToast.show();
        
        toast.addEventListener('hidden.bs.toast', function() {
            toast.remove();
        });
    }
    
    // Handle light/dark mode preference
    const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)');
    if (prefersDarkScheme.matches) {
        document.body.classList.add('dark-mode');
    }
}); 