document.addEventListener('DOMContentLoaded', function() {
    const addFeedForm = document.getElementById('add-feed-form');
    const feedUrlInput = document.getElementById('feed-url');
    const feedList = document.getElementById('feed-list');
    const feedContent = document.getElementById('feed-content');
    const currentFeedTitle = document.getElementById('current-feed-title');

    // Add a new feed
    addFeedForm.addEventListener('submit', function(e) {
        e.preventDefault();
        const url = feedUrlInput.value.trim();
        
        if (!url) {
            showToast('Please enter a valid URL', 'danger');
            return;
        }

        axios.post('/feeds', { url: url })
            .then(function(response) {
                feedUrlInput.value = '';
                showToast('Feed added successfully!', 'success');
                
                // Add the new feed to the list
                const feed = response.data;
                addFeedToList(feed);
            })
            .catch(function(error) {
                showToast('Error adding feed: ' + (error.response?.data?.error || error.message), 'danger');
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
            
            loadFeed(url);
        }
        
        // Delete feed when delete button is clicked
        if (e.target.classList.contains('delete-feed')) {
            const url = e.target.dataset.url;
            deleteFeed(url);
            e.stopPropagation();
        }
    });

    // Load feed content - Updated to use query parameters
    function loadFeed(url) {
        axios.get('/feed', { params: { url: url } })
            .then(function(response) {
                const feed = response.data;
                currentFeedTitle.textContent = feed.Title;
                
                let html = '';
                if (feed.Items && feed.Items.length > 0) {
                    feed.Items.forEach(item => {
                        const date = new Date(item.PublishedAt);
                        const formattedDate = date.toLocaleString();
                        
                        html += `
                            <div class="feed-item-entry">
                                <h5><a href="${item.Link}" target="_blank">${item.Title}</a></h5>
                                <div class="feed-item-meta">
                                    <span>${formattedDate}</span>
                                </div>
                                <div class="feed-item-content">
                                    ${item.Description || item.Content || ''}
                                </div>
                            </div>
                        `;
                    });
                } else {
                    html = '<p class="text-center text-muted">No items found in this feed</p>';
                }
                
                feedContent.innerHTML = html;
            })
            .catch(function(error) {
                showToast('Error loading feed: ' + (error.response?.data?.error || error.message), 'danger');
                feedContent.innerHTML = '<p class="text-center text-danger">Error loading feed</p>';
            });
    }

    // Delete a feed - Updated to use query parameters
    function deleteFeed(url) {
        if (confirm('Are you sure you want to delete this feed?')) {
            axios.delete('/feed', { params: { url: url } })
                .then(function() {
                    // Remove the feed from the DOM
                    const feedItem = document.querySelector(`.feed-item[data-url="${url}"]`);
                    if (feedItem) {
                        feedItem.remove();
                    }
                    
                    // Clear feed content if it was the active feed
                    if (feedItem && feedItem.classList.contains('active')) {
                        currentFeedTitle.textContent = 'Select a Feed';
                        feedContent.innerHTML = '<p class="text-center text-muted">Select a feed from the list to view its contents</p>';
                    }
                    
                    showToast('Feed deleted successfully!', 'success');
                })
                .catch(function(error) {
                    showToast('Error deleting feed: ' + (error.response?.data?.error || error.message), 'danger');
                });
        }
    }

    // Add a feed to the list - Updated to use query parameters
    function addFeedToList(feed) {
        const li = document.createElement('li');
        li.className = 'list-group-item d-flex justify-content-between align-items-center feed-item';
        li.dataset.url = feed.URL;
        li.innerHTML = `
            <span class="feed-title">${feed.Title}</span>
            <div>
                <a href="/export?url=${encodeURIComponent(feed.URL)}" class="btn btn-sm btn-outline-secondary" target="_blank">RSS</a>
                <button class="btn btn-sm btn-danger delete-feed" data-url="${feed.URL}">Delete</button>
            </div>
        `;
        feedList.appendChild(li);
    }

    // Show toast notification
    function showToast(message, type) {
        const toast = document.createElement('div');
        toast.className = `toast bg-${type} text-white`;
        toast.setAttribute('role', 'alert');
        toast.setAttribute('aria-live', 'assertive');
        toast.setAttribute('aria-atomic', 'true');
        toast.innerHTML = `
            <div class="toast-body">
                ${message}
            </div>
        `;
        
        document.body.appendChild(toast);
        
        const bsToast = new bootstrap.Toast(toast, {
            autohide: true,
            delay: 3000
        });
        
        bsToast.show();
        
        toast.addEventListener('hidden.bs.toast', function() {
            toast.remove();
        });
    }
}); 