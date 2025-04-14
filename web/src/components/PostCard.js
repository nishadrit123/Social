import React, { useState } from "react";
import "bootstrap/dist/css/bootstrap.min.css";
import {
  FaBookmark,
  FaRegBookmark,
  FaHeart,
  FaRegHeart,
  FaComment,
} from "react-icons/fa";
import axios from "axios";

const PostCard = ({ post }) => {
  const {
    id,
    title,
    content,
    tags,
    created_at,
    user,
    like_count,
    comment_count,
    is_post_liked,
    is_post_saved,
  } = post;

  // Local UI state
  const [isLiked, setIsLiked] = useState(is_post_liked || false);
  const [likeCount, setLikeCount] = useState(Number(like_count) || 0);
  const [isSaved, setIsSaved] = useState(post.is_post_saved);

  const handleLike = async () => {
    // Optimistic UI update
    const updatedLiked = !isLiked;
    setIsLiked(updatedLiked);
    setLikeCount((prev) => (updatedLiked ? prev + 1 : prev - 1));

    try {
      const token = localStorage.getItem("jwtToken");
      await axios.post(
        `http://localhost:8080/v1/likedislike/post/${id}/`,
        {},
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );
    } catch (error) {
      console.error("Like API failed, reverting UI changes...");
      // Revert UI on failure
      setIsLiked((prev) => !prev);
      setLikeCount((prev) => (updatedLiked ? prev - 1 : prev + 1));
    }
  };

  const handleSaveToggle = async () => {
    try {
      const token = localStorage.getItem("jwtToken");
      // Optimistically toggle UI
      setIsSaved((prev) => !prev);

      await axios.post(
        `http://localhost:8080/v1/posts/${post.id}/saveunsave`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
    } catch (error) {
      console.error("Error saving/un-saving post:", error);
      // Rollback if needed
      setIsSaved((prev) => !prev);
    }
  };

  return (
    <div className="card mb-4">
      <div className="card-body">
        {/* Username and Date */}
        <div className="d-flex justify-content-between">
          <span className="text-muted">
            <strong>@{user?.username || "anonymous"}</strong>
          </span>
          <span className="text-muted">
            {new Date(created_at).toLocaleDateString()}
          </span>
        </div>

        <h5 className="card-title mt-2">{title}</h5>
        <p className="card-text">{content}</p>

        <p className="text-muted">
          <strong>Tags:</strong> {tags?.join(", ")}
        </p>

        <div className="d-flex justify-content-between align-items-center mt-3">
          {/* ❤️ Like Count */}
          <span
            className="d-flex align-items-center gap-2"
            style={{ cursor: "pointer" }}
            onClick={handleLike}
          >
            {isLiked ? <FaHeart color="red" /> : <FaRegHeart color="gray" />}
            {likeCount}
          </span>

          {/* 💬 Comment Count */}
          <span className="d-flex align-items-center gap-2">
            <FaComment /> {comment_count}
          </span>

          {/* 🔖 Save Icon */}
          <span
            className="d-flex align-items-center gap-2"
            onClick={handleSaveToggle}
            style={{ cursor: "pointer" }}
          >
            {isSaved ? (
              <FaBookmark color="black" />
            ) : (
              <FaRegBookmark color="gray" />
            )}
          </span>
        </div>
      </div>
    </div>
  );
};

export default PostCard;
