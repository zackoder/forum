let offset = 0;
const limit = 20;
let loading = false;

async function loadMorePosts(path) {
  if (loading) return;
  loading = true;
  try {
    const response = await fetch(`/${path}?offset=${offset}`);

    const posts = await response.json();
    if (!posts || posts.length === 0) return;
    console.log(typeof posts);
    createPosts(posts);

    offset += limit;
  } catch (error) {
    console.error("Error loading posts:", error);
  } finally {
    loading = false;
  }
}

function createPosts(posts) {
  const postsContainer = document.getElementById("posts-container");
  posts.forEach((post) => {
    const postElement = document.createElement("div");
    postElement.className = "post-container";
    postElement.dataset.postId = post.ID;

    /* h2 will contain the image and name of the persen who posted */
    const posterName = createEle("h2");
    posterName.className = "poster";
    const posterImg = createEle("img");
    posterImg.src = "/css/466006304_871124095226532_8631138819273739648_n.jpg";
    const nameContainer = createEle("span");
    nameContainer.className = "UserName";
    nameContainer.innerText = post.UserName;
    posterName.appendChild(posterImg);
    posterName.appendChild(nameContainer);
    postElement.appendChild(posterName);

    /* creating a div that will contain all the elements bellow */
    const pc = createEle("div");
    pc.className = "pc";

    /* creating an h3 element to contain the post title */
    const title = createEle("h3");
    title.className = "title";
    title.innerText = post.Title;

    /* creating a p element that will contain the content of the post */
    const content = createEle("p");
    content.className = "content";
    content.innerText = post.Content;
    pc.append(title, content);

    /* creating like and dislike button */
    const like_dislike_container = createEle("div");
    like_dislike_container.className = "like-dislike-container";

    /* creating of the like button */
    const likebnt = createEle("button");
    likebnt.className = "like-btn";

    /* create an img element to contain like icon */
    const likeIcon = createEle("img");
    likeIcon.className = "likeicon";
    likeIcon.src = "/css/like.png";
    likebnt.appendChild(likeIcon);

    /* creationg of the dislike button */
    const dislikebnt = createEle("button");
    dislikebnt.className = "dislike-btn";

    /* creating an img tag to containg dislike icon */
    const dislikeIcone = createEle("img");
    dislikeIcone.className = "dislikeicon";
    dislikeIcone.src = "/css/dislike.png";

    dislikebnt.appendChild(dislikeIcone);

    /* appending like and dislike buttons to like container */
    like_dislike_container.append(likebnt, dislikebnt);

    /* appending like container to the post contaner */
    pc.appendChild(like_dislike_container);

    /* adding a button to see comments */
    const seecomments = createEle("button");
    seecomments.className = "see_comments";
    seecomments.innerText = "see comments";
    pc.appendChild(seecomments);
    /* creating the form that sends comments */
    const comment_form = createEle("form");
    comment_form.method = "POST";
    comment_form.className = "comment_form";

    const title_impt = createEle("input");
    title_impt.className = "comment";
    title_impt.name = "comment";
    title_impt.type = "text";
    title_impt.placeholder = "Add your comment";
    title_impt.required = true;

    const submit_comment = createEle("button");
    submit_comment.className = "send_comment";
    submit_comment.type = "submit";

    const send_icon = createEle("img");
    send_icon.className = "sendimg";
    send_icon.src = "/css/send-message.png";
    submit_comment.appendChild(send_icon);
    comment_form.appendChild(title_impt);
    comment_form.appendChild(submit_comment);

    pc.appendChild(comment_form);
    postElement.appendChild(pc);

    postsContainer.appendChild(postElement);
  });
}

function createEle(elename) {
  return document.createElement(elename);
}

// function handleScroll(path) {
//   console.log(path);
//   const scrollPosition = window.scrollY + window.innerHeight;
//   const threshold = document.body.scrollHeight - 1000;

//   if (scrollPosition > threshold) {
//     loadMorePosts(path);
//   }
// }
async function addEventOnPosts(path) {
  document.addEventListener("DOMContentLoaded", function () {
    const postsContainer = document.getElementById("posts-container");

    // Event delegation for click events
    postsContainer.addEventListener("click", function (event) {
      const postElement = event.target.closest(".post-container");
      if (!postElement) return;

      const postId = postElement.getAttribute("data-post-id");

      if (
        event.target.classList.contains("like-btn") ||
        event.target.classList.contains("likeicon")
      ) {
        alert("action");
        handleLike(postId, true);
      } else if (
        event.target.classList.contains("dislike-btn") ||
        event.target.classList.contains("dislikeicon")
      ) {
        handleLike(postId, false);
      }
    });

    postsContainer.addEventListener("submit", function (event) {
      const postElement = event.target.closest(".post-container");
      if (event.target.classList.contains("comment_form")) {
        event.preventDefault();
        const form = event.target;

        const postId = postElement.getAttribute("data-post-id");
        const commentText = form.querySelector(".comment").value.trim();

        if (commentText === "") {
          alert("Comment cannot be empty.");
          return;
        }

        handleComment(postId, commentText);
        form.reset();
      }
    });

    loadMorePosts(path);
    
  });
}

function handleLike(postId, like) {
  fetch("/like-post", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body: `post_id=${postId}&like=${like}`,
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return response.json();
    })
    .then((data) => {
      console.log("Like/Dislike updated:", data);
    })
    .catch((error) => console.error("Error updating like/dislike:", error));
}
