import React from 'react';
import "bootstrap/dist/css/bootstrap.min.css";

const PostCard = ({ post }) => {
  const {
    title,
    content,
    tags,
    created_at,
    user,
    like_count,
    comment_count,
  } = post;

  const displayUsername = user?.username || "Unknown";
  const displayTags = tags?.length ? tags.join(', ') : 'No tags';

  return (
    <div className="card mb-4 shadow-sm">
      <div className="card-header d-flex justify-content-between align-items-center">
        <strong>@{displayUsername}</strong>
        <small>{new Date(created_at).toLocaleDateString()}</small>
      </div>
      <div className="card-body">
        <h5 className="card-title">{title || 'No title'}</h5>
        <p className="card-text">{content}</p>
        <p><strong>Tags:</strong> {displayTags}</p>
      </div>
      <div className="card-footer d-flex justify-content-between">
        <span>❤️ {like_count || 0}</span>
        <span>💬 {comment_count || 0}</span>
      </div>
    </div>
  );
};

export default PostCard;
