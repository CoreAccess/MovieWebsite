/**
 * StreamLine Main JavaScript
 * Handles mock data, dynamic interactions, and search page filtering.
 */

document.addEventListener('DOMContentLoaded', () => {
    // Initialize tooltips and popovers if necessary
    const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]')
    const tooltipList = [...tooltipTriggerList].map(tooltipTriggerEl => new bootstrap.Tooltip(tooltipTriggerEl))

    console.log("StreamLine UI Initialized.");
    
    // Initialize horizontal feed scroll visibility
    initFeedScrollIndicators();
});

/**
 * Scrolls a horizontal feed container
 * @param {HTMLElement} btn - The button that was clicked
 * @param {number} direction - 1 for right, -1 for left
 */
function scrollFeed(btn, direction) {
    const wrapper = btn.closest('.feed-scroll-wrapper');
    const container = wrapper.querySelector('.feed-container');
    
    if (container) {
        // Scroll by roughly the visible width of the container
        const scrollAmount = container.clientWidth * 0.8 * direction;
        container.scrollBy({ left: scrollAmount, behavior: 'smooth' });
    }
}

/**
 * Initializes and manages the visibility of left/right scroll buttons
 * on horizontal feeds.
 */
function initFeedScrollIndicators() {
    const wrappers = document.querySelectorAll('.feed-scroll-wrapper');
    
    wrappers.forEach(wrapper => {
        const container = wrapper.querySelector('.feed-container');
        const leftBtn = wrapper.querySelector('.feed-scroll-btn-left');
        const rightBtn = wrapper.querySelector('.feed-scroll-btn-right');
        
        if (!container || !leftBtn || !rightBtn) return;
        
        const updateButtons = () => {
            // Check if we are at the start
            if (container.scrollLeft <= 10) {
                leftBtn.style.opacity = '0';
                leftBtn.style.pointerEvents = 'none';
            } else {
                leftBtn.style.opacity = '0.9';
                leftBtn.style.pointerEvents = 'auto';
            }
            
            // Check if we are at the end
            // scrollWidth is total width, clientWidth is visible width
            // math.ceil and a small buffer (10px) to handle fractional pixels
            if (Math.ceil(container.scrollLeft + container.clientWidth) >= container.scrollWidth - 10) {
                rightBtn.style.opacity = '0';
                rightBtn.style.pointerEvents = 'none';
            } else {
                rightBtn.style.opacity = '0.9';
                rightBtn.style.pointerEvents = 'auto';
            }
        };
        
        // Initial check
        updateButtons();
        
        // Update on scroll
        container.addEventListener('scroll', updateButtons, { passive: true });
        
        // Update on window resize since it changes container dims
        window.addEventListener('resize', updateButtons, { passive: true });
    });
}
