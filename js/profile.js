let path = "profile";
loadMorePosts(path);

addEventOnPosts(path);

// window.addEventListener("scroll", _.throttle(handleScroll, 500));

// function handleScroll() {
//   const scrollPosition = window.scrollY + window.innerHeight;
//   const threshold = document.body.scrollHeight - 1000;

//   if (scrollPosition > threshold) {
//     loadMorePosts(path);
//   }
// }
