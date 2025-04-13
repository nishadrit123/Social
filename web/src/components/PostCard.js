import React from "react";
import "bootstrap/dist/css/bootstrap.min.css";
import { FaHeart, FaRegHeart, FaComment } from "react-icons/fa";
import { FaRegBookmark, FaBookmark } from "react-icons/fa6"; // Save Icons

const PostCard = ({ post }) => {
  const {
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

  return (
    <div className="card mb-4">
      <div className="card-body">
        {/* Username top left like Instagram */}
        <div className="d-flex justify-content-between align-items-center mb-3">
          <h6 className="text-muted m-0">
            <strong>@{user?.username || "anonymous"}</strong>
          </h6>
          <span className="text-muted" style={{ fontSize: "0.85rem" }}>
            {new Date(created_at).toLocaleDateString()}
          </span>
        </div>

        <h5 className="card-title">{title}</h5>
        <p className="card-text">{content}</p>

        <p className="text-muted">
          <strong>Tags:</strong> {tags?.join(", ")}
        </p>

        <div className="d-flex justify-content-between align-items-center mt-3">
          {/* ❤️ Like Count */}
          <span className="d-flex align-items-center gap-2">
            {is_post_liked ? (
              <FaHeart color="red" />
            ) : (
              <FaRegHeart color="gray" />
            )}
            {like_count}
          </span>

          {/* 💬 Comment Count */}
          <span className="d-flex align-items-center gap-2">
            <FaComment /> {comment_count}
          </span>

          {/* 🔖 Save Icon */}
          <span className="d-flex align-items-center">
            {is_post_saved ? (
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
