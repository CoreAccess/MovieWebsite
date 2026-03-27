/**
 * home.js - Homepage specific interactions
 * Handles Trending Matrix filters and horizontal scrolling.
 */

document.addEventListener("DOMContentLoaded", () => {
	initTrendingMatrix();
});

function initTrendingMatrix() {
	const scrollContainer = document.getElementById("trending-scroll");
	const filtersContainer = document.getElementById("trending-filters");
	const btnPrev = document.getElementById("trending-prev");
	const btnNext = document.getElementById("trending-next");

	if (!scrollContainer || !filtersContainer) return;

	// 1. Filter Handling
	filtersContainer.addEventListener("click", async (e) => {
		const btn = e.target.closest("button");
		if (!btn) return;

		// Update UI state
		filtersContainer.querySelectorAll("button").forEach((b) => {
			b.classList.remove("active", "btn-white", "shadow-sm");
			b.classList.add("btn-transparent");
		});
		btn.classList.add("active", "btn-white", "shadow-sm");
		btn.classList.remove("btn-transparent");

		const type = btn.dataset.type;
		await fetchTrending(type);
	});

	// 2. Horizontal Scrolling
	if (btnPrev && btnNext) {
		btnPrev.addEventListener("click", () => {
			scrollContainer.scrollBy({ left: -400, behavior: "smooth" });
		});
		btnNext.addEventListener("click", () => {
			scrollContainer.scrollBy({ left: 400, behavior: "smooth" });
		});

		// Hide/Show buttons based on scroll position
		const updateScrollButtons = () => {
			const isAtStart = scrollContainer.scrollLeft <= 0;
			const isAtEnd = scrollContainer.scrollLeft + scrollContainer.clientWidth >= scrollContainer.scrollWidth - 1;

			btnPrev.style.display = isAtStart ? "none" : "flex";
			btnNext.style.display = isAtEnd ? "none" : "flex";
		};

		scrollContainer.addEventListener("scroll", updateScrollButtons);
		window.addEventListener("resize", updateScrollButtons);
		updateScrollButtons(); // Initial check
	}

	async function fetchTrending(type) {
		try {
			// Show loading state (optional)
			scrollContainer.style.opacity = "0.5";

			const response = await fetch(`/api/trending?type=${type}`);
			if (!response.ok) throw new Error("Network response was not ok");

			const data = await response.json();
			renderTrending(data);
		} catch (err) {
			console.error("Failed to fetch trending:", err);
		} finally {
			scrollContainer.style.opacity = "1";
		}
	}

	function renderTrending(items) {
		scrollContainer.replaceChildren();

		if (!items || items.length === 0) {
			const noItems = document.createElement("div");
			noItems.className = "p-5 text-center w-100 text-muted";
			noItems.textContent = "No items found";
			scrollContainer.appendChild(noItems);
			return;
		}

		const fragment = document.createDocumentFragment();

		items.forEach((item) => {
			// Determine the URL prefix based on media type
			let prefix = "movies";
			if (item.media_type === "TV" || item.media_type === "TVSeries") prefix = "tv-shows";
			if (item.media_type === "People") prefix = "people";

			// For People: show age if available. For media: show year + rating.
			let subInfo = "";
			if (item.media_type === "People") {
				subInfo = item.year ? `Age: ${item.year}` : "";
			} else {
				const rating = (item.aggregate_rating || 0).toFixed(1);
				subInfo = item.year ? `${item.year} • ${rating}` : rating;
			}

			const wrapper = document.createElement("div");
			// Replaced inline style with a class. Make sure to define .trending-item { width: 180px; } in style.css
			wrapper.className = "trending-item flex-shrink-0";

			const link = document.createElement("a");
			link.href = `/${prefix}/${item.id}/${item.slug}`;
			link.className = "text-decoration-none";

			const img = document.createElement("img");
			img.src = item.image;
			img.className = "rounded shadow-sm mb-2 w-100 object-fit-cover aspect-poster";
			img.alt = item.name;

			const title = document.createElement("h6");
			title.className = "text-dark fw-bold mb-0 text-truncate";
			title.textContent = item.name;

			const infoSpan = document.createElement("span");
			infoSpan.className = "text-muted small";
			infoSpan.textContent = subInfo;

			link.appendChild(img);
			link.appendChild(title);
			link.appendChild(infoSpan);
			wrapper.appendChild(link);

			fragment.appendChild(wrapper);
		});

		scrollContainer.appendChild(fragment);
	}
}
