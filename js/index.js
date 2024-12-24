function handleComment(postId, comment) {
  fetch("/comments", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body: `post_id=${postId}&comment=${comment}`,
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return response.json();
    })
    .then((data) => {
      if (data.success) {
        alert("Comment added successfully!");
      } else {
        alert("Failed to add comment.");
      }
    })
    .catch((error) => console.error("Error submitting comment:", error));
}

let profile = document.getElementById("profile");
if (!profile) {
  document.getElementById("posts-container").style.paddingTop = "120px";
}

const showPostFormButton = document.querySelector(".show-postForm");
const postForm = document.querySelector(".postForm");
const layout = document.querySelector(".lay-out");
// Show the form
showPostFormButton.addEventListener("click", () => {
  postForm.style.display = "flex";
  layout.style.display = "block";
  document.body.style.overflow = "hidden";
});

// Hide the form when clicking outside or on a cancel button
layout.addEventListener("click", () => {
  postForm.style.display = "none";
  layout.style.display = "none";
  document.body.style.overflow = "";
});
