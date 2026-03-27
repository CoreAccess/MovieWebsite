/**
 * FilmGap Main JavaScript — ES6 Module
 * Handles dynamic interactions, horizontal feeds, and user interaction modals.
 * No global functions or variables — all logic is encapsulated within this module.
 */

document.addEventListener("DOMContentLoaded", () => {
	// Initialize Bootstrap tooltips
	const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]');
	[...tooltipTriggerList].map((el) => new bootstrap.Tooltip(el));

	initFeedScrollIndicators();
	initFeedScrollDelegation();
	initInteractionModal();
	initTabStateSync();
	initGlobalSubmitHandler();
});

/**
 * Initializes the Unified Interaction Modal.
 * Populates media context (ID, type, name) from the triggering button's data attributes.
 */
function initInteractionModal() {
	const modal = document.getElementById("interactionModal");
	if (!modal) return;

	modal.addEventListener("show.bs.modal", (event) => {
		const button = event.relatedTarget;
		const mediaId = button.getAttribute("data-media-id");
		const mediaType = button.getAttribute("data-media-type");
		const mediaName = button.getAttribute("data-media-name");

		modal.querySelector("#modalMediaId").value = mediaId;
		modal.querySelector("#modalMediaType").value = mediaType;
		modal.querySelector("#interactionModalLabel").textContent = `Log ${mediaName}`;

		modal.querySelector("#interactionForm").reset();
	});

	// Watchlist toggle via AJAX — no page reload, just toggle the button state
	const watchlistBtn = modal.querySelector("#modalWatchlistBtn");
	if (watchlistBtn) {
		watchlistBtn.addEventListener("click", async () => {
			const mediaId = modal.querySelector("#modalMediaId").value;
			const mediaType = modal.querySelector("#modalMediaType").value;
			const csrfToken = modal.querySelector('input[name="csrf_token"]').value;
			const isAdded = watchlistBtn.dataset.added === "true";

			const formData = new FormData();
			formData.append("media_id", mediaId);
			formData.append("media_type", mediaType);
			formData.append("action", isAdded ? "remove" : "add");
			formData.append("csrf_token", csrfToken);

			try {
				const response = await fetch("/watchlist/toggle", {
					method: "POST",
					body: formData,
				});
				if (response.ok) {
					const icon = watchlistBtn.querySelector("i");
					if (isAdded) {
						icon.className = "bi bi-bookmark-plus";
						watchlistBtn.classList.remove("btn-success");
						watchlistBtn.classList.add("btn-outline-dark");
						watchlistBtn.dataset.added = "false";
					} else {
						icon.className = "bi bi-bookmark-check-fill";
						watchlistBtn.classList.remove("btn-outline-dark");
						watchlistBtn.classList.add("btn-success");
						watchlistBtn.dataset.added = "true";
					}
				}
			} catch (err) {
				console.error("Watchlist toggle failed:", err);
			}
		});
	}
}

/**
 * Uses event delegation to handle horizontal feed scroll button clicks.
 * Replaces the old global scrollFeed() function — no inline onclick attributes needed.
 * Buttons must have class .feed-scroll-btn-left or .feed-scroll-btn-right
 * and be nested inside a .feed-scroll-wrapper that contains a .feed-container.
 */
function initFeedScrollDelegation() {
	document.addEventListener("click", (e) => {
		const leftBtn = e.target.closest(".feed-scroll-btn-left");
		const rightBtn = e.target.closest(".feed-scroll-btn-right");

		const btn = leftBtn || rightBtn;
		if (!btn) return;

		const direction = leftBtn ? -1 : 1;
		const wrapper = btn.closest(".feed-scroll-wrapper");
		if (!wrapper) return;

		const container = wrapper.querySelector(".feed-container");
		if (!container) return;

		const scrollAmount = container.clientWidth * 0.8 * direction;
		container.scrollBy({ left: scrollAmount, behavior: "smooth" });
	});
}

/**
 * Initializes and manages the visibility of left/right scroll buttons
 * on horizontal feeds based on current scroll position.
 */
function initFeedScrollIndicators() {
	const wrappers = document.querySelectorAll(".feed-scroll-wrapper");

	wrappers.forEach((wrapper) => {
		const container = wrapper.querySelector(".feed-container");
		const leftBtn = wrapper.querySelector(".feed-scroll-btn-left");
		const rightBtn = wrapper.querySelector(".feed-scroll-btn-right");

		if (!container || !leftBtn || !rightBtn) return;

		const updateButtons = () => {
			const atStart = container.scrollLeft <= 10;
			const atEnd = Math.ceil(container.scrollLeft + container.clientWidth) >= container.scrollWidth - 10;

			leftBtn.style.opacity = atStart ? "0" : "0.9";
			leftBtn.style.pointerEvents = atStart ? "none" : "auto";
			rightBtn.style.opacity = atEnd ? "0" : "0.9";
			rightBtn.style.pointerEvents = atEnd ? "none" : "auto";
		};

		updateButtons();
		container.addEventListener("scroll", updateButtons, { passive: true });
		window.addEventListener("resize", updateButtons, { passive: true });
	});
}

function initTabStateSync() {
	document.addEventListener("shown.bs.tab", (event) => {
		const activeTab = event.target;
		const tabList = activeTab.closest('[role="tablist"]');
		if (!tabList) return;

		const isReferenceTabBar = tabList.classList.contains("ref-tabs");
		tabList.querySelectorAll('[role="tab"]').forEach((tab) => {
			const isActive = tab === activeTab;
			tab.classList.toggle("active", isActive);
			if (isReferenceTabBar) {
				tab.classList.remove("bg-subtle-primary", "text-primary-accent", "text-muted", "fw-medium", "shadow-sm");
				tab.classList.toggle("fw-bold", isActive);
				return;
			}

			tab.classList.toggle("bg-subtle-primary", isActive);
			tab.classList.toggle("text-primary-accent", isActive);
			tab.classList.toggle("fw-bold", isActive);
			tab.classList.toggle("shadow-sm", isActive);
			tab.classList.toggle("text-muted", !isActive);
			tab.classList.toggle("fw-medium", !isActive);
		});
	});
}

/**
 * Global submit handler for AJAX forms (Follow, Watchlist, etc.)
 */
function initGlobalSubmitHandler() {
	document.addEventListener("submit", async (e) => {
		const followForm = e.target.closest(".follow-form");
		const watchlistForm = e.target.closest(".watchlist-form");

		if (!followForm && !watchlistForm) return;

		e.preventDefault();
		const form = followForm || watchlistForm;
		const formData = new FormData(form);
		const url = form.getAttribute("action") || "/watchlist/toggle";
		const btn = form.querySelector("button");

		// Save state for rollback without innerHTML
		const originalChildren = Array.from(btn.childNodes).map((n) => n.cloneNode(true));
		const restoreOriginalState = () => btn.replaceChildren(...originalChildren);

		// Helper to safely build button content
		const updateBtnContent = (iconClass, text) => {
			btn.replaceChildren();
			if (iconClass) {
				const icon = document.createElement("i");
				icon.className = iconClass;
				btn.appendChild(icon);
			}
			btn.appendChild(document.createTextNode((iconClass ? " " : "") + text));
		};

		// State detection for watchlist
		if (watchlistForm && !formData.has("action")) {
			const isAdded = btn.textContent.includes("On Watchlist") || btn.querySelector(".bi-bookmark-check-fill");
			formData.append("action", isAdded ? "remove" : "add");
		}

		// Feedback
		btn.disabled = true;

		try {
			const response = await fetch(url, {
				method: "POST",
				body: formData,
				headers: {
					"X-Requested-With": "XMLHttpRequest",
				},
			});

			if (response.status === 401) {
				window.location.href = "/login?next=" + encodeURIComponent(window.location.pathname);
				return;
			}

			if (response.ok) {
				const data = await response.json();

				if (followForm) {
					if (data.success) {
						const isUnfollow = url.includes("unfollow");
						if (isUnfollow) {
							form.setAttribute("action", "/follow");
							btn.className = "btn btn-primary-accent rounded-1 fw-bold shadow-sm px-4";
							updateBtnContent("bi bi-person-plus me-1", "Follow");
						} else {
							form.setAttribute("action", "/unfollow");
							btn.className = "btn btn-outline-primary rounded-1 fw-bold shadow-sm px-4";
							updateBtnContent("bi bi-person-dash me-1", "Unfollow");
						}
					}
				} else if (watchlistForm) {
					if (data.added) {
						updateBtnContent("bi bi-bookmark-check-fill me-2", "On Watchlist");
					} else {
						if (btn.classList.contains("px-4") || btn.classList.contains("py-2")) {
							updateBtnContent("bi bi-bookmark-plus me-2", "Watchlist");
						} else {
							updateBtnContent("", "+ Add to Watchlist");
						}
					}
				}
			} else {
				console.error("Action failed");
				restoreOriginalState();
			}
		} catch (err) {
			console.error("Action error:", err);
			restoreOriginalState();
		} finally {
			btn.disabled = false;
		}
	});
}
