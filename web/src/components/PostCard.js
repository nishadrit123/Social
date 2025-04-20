import React, { useState } from "react";
import "bootstrap/dist/css/bootstrap.min.css";
import {
  FaBookmark,
  FaRegBookmark,
  FaHeart,
  FaRegHeart,
  FaComment,
  FaEdit,
  FaTrash,
} from "react-icons/fa";
import axios from "axios";
import LikedUsersModal from "./LikedUsersModal";
import CommentsModal from "./CommentsModal";
import { useNavigate } from "react-router-dom";
import "../style/PostCard.css";

const PostCard = ({ post, isOwnProfile, onEdit, onDelete }) => {
  const {
    id,
    title,
    content,
    tags,
    created_at,
    user,
    like_count,
    is_post_liked,
    is_post_saved,
  } = post;

  const [isLiked, setIsLiked] = useState(is_post_liked || false);
  const [likeCount, setLikeCount] = useState(Number(like_count) || 0);
  const [commentCount, setCommentCount] = useState(
    Number(post.comment_count) || 0
  );
  const decrementCommentCount = () => {
    setCommentCount((prev) => Math.max(prev - 1, 0));
  };
  const [isSaved, setIsSaved] = useState(is_post_saved || false);

  const [showModal, setShowModal] = useState(false);
  const [likedUsers, setLikedUsers] = useState([]);

  const [showCommentModal, setShowCommentModal] = useState(false);
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState("");

  const [editingCommentId, setEditingCommentId] = useState(null);
  const [editText, setEditText] = useState("");
  const [isHovered, setIsHovered] = useState(false);
  const navigate = useNavigate();

  const handleLike = async () => {
    const updatedLiked = !isLiked;
    setIsLiked(updatedLiked);
    setLikeCount((prev) => (updatedLiked ? prev + 1 : prev - 1));

    try {
      const token = localStorage.getItem("jwtToken");
      await axios.post(
        `http://localhost:8080/v1/likedislike/post/${id}/`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
    } catch (error) {
      console.error("Like API failed, reverting UI changes...");
      setIsLiked((prev) => !prev);
      setLikeCount((prev) => (updatedLiked ? prev - 1 : prev + 1));
    }
  };

  const handleLikeCountClick = async () => {
    try {
      const token = localStorage.getItem("jwtToken");
      const res = await axios.get(
        `http://localhost:8080/v1/likedislike/post/${id}/`,
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setLikedUsers(res.data.data || []);
      setShowModal(true);
    } catch (error) {
      console.error("Error fetching liked users", error);
    }
  };

  const handleSaveToggle = async () => {
    try {
      const token = localStorage.getItem("jwtToken");
      setIsSaved((prev) => !prev);
      await axios.post(
        `http://localhost:8080/v1/posts/${id}/saveunsave`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
    } catch (error) {
      console.error("Save API failed, reverting UI changes...");
      setIsSaved((prev) => !prev);
    }
  };

  const handleOpenCommentModal = async () => {
    try {
      const token = localStorage.getItem("jwtToken");
      const res = await axios.get(
        `http://localhost:8080/v1/comment/post/${id}/`,
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setComments(res.data.data || []);
      setShowCommentModal(true);
    } catch (err) {
      console.error("Failed to fetch comments", err);
    }
  };

  const handlePostComment = async () => {
    if (!newComment.trim()) return;

    const commentToPost = newComment;
    setNewComment("");
    setCommentCount((prev) => prev + 1);

    try {
      const token = localStorage.getItem("jwtToken");
      await axios.post(
        `http://localhost:8080/v1/comment/post/${id}/`,
        { comment: commentToPost },
        { headers: { Authorization: `Bearer ${token}` } }
      );

      setComments((prev) => [
        ...prev,
        { username: "You", comment: commentToPost },
      ]);
    } catch (error) {
      console.error("Failed to post comment", error);
      setNewComment(commentToPost);
    }
  };

  const handleDeleteComment = async (commentObj) => {
    const confirmDelete = window.confirm(
      "Are you sure you want to delete this comment?"
    );
    if (!confirmDelete) return;

    try {
      const token = localStorage.getItem("jwtToken");
      await axios.delete(`http://localhost:8080/v1/comment/${commentObj.id}/`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      // Remove from UI
      setComments((prev) => prev.filter((c) => c.id !== commentObj.id));
      decrementCommentCount();
    } catch (error) {
      console.error("Failed to delete comment", error);
      alert("Could not delete the comment. Please try again.");
    }
  };

  const handleEditClick = (commentObj) => {
    setEditingCommentId(commentObj.id);
    setEditText(commentObj.comment);
  };

  const handleCancelEdit = () => {
    setEditingCommentId(null);
    setEditText("");
  };

  const handleSaveEdit = async (commentObj) => {
    try {
      const token = localStorage.getItem("jwtToken");
      await axios.patch(
        `http://localhost:8080/v1/comment/${commentObj.id}/`,
        { comment: editText },
        { headers: { Authorization: `Bearer ${token}` } }
      );

      setComments((prev) =>
        prev.map((c) =>
          c.id === commentObj.id ? { ...c, comment: editText } : c
        )
      );
      setEditingCommentId(null);
      setEditText("");
    } catch (err) {
      console.error("Failed to edit comment", err);
    }
  };

  const handleUserClick = () => {
    navigate(`/profile/${post.user_id}`);
  };

  return (
    <>
      <div
        className="card mb-4"
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        {isOwnProfile && isHovered && (
          <div className="post-actions">
            <button
              className="action-button delete"
              onClick={() => onDelete(post.id)}
              aria-label="Delete Post"
              title="Delete Post"
            >
              <FaTrash />
            </button>
            <button
              className="action-button"
              onClick={() => onEdit(post)}
              aria-label="Edit Post"
              title="Edit Post"
            >
              <FaEdit />
            </button>
          </div>
        )}
        <div className="card-body" style={{width: "300px"}}>
          <div className="d-flex justify-content-between">
            <span className="text-muted">
              <strong
                onClick={handleUserClick}
                style={{
                  cursor: "pointer",
                  fontWeight: "bold",
                  color: "#007bff",
                }}
              >
                @{user?.username || "anonymous"}
              </strong>
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
            {/* ❤️ Like */}
            <span className="d-flex align-items-center gap-2">
              <span style={{ cursor: "pointer" }} onClick={handleLike}>
                {isLiked ? (
                  <FaHeart color="red" />
                ) : (
                  <FaRegHeart color="gray" />
                )}
              </span>
              <span
                style={{ cursor: "pointer" }}
                onClick={handleLikeCountClick}
                className="text-primary"
              >
                {likeCount}
              </span>
            </span>

            {/* 💬 Comment */}
            <span
              className="d-flex align-items-center gap-2"
              style={{ cursor: "pointer" }}
              onClick={handleOpenCommentModal}
            >
              <FaComment /> {commentCount}
            </span>

            {/* 🔖 Save */}
            <span
              onClick={handleSaveToggle}
              style={{ cursor: "pointer" }}
              className="d-flex align-items-center"
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

      <LikedUsersModal
        show={showModal}
        onClose={() => setShowModal(false)}
        likedUsers={likedUsers}
        title={"Liked by"}
        emptytitle={"likes"}
      />

      <CommentsModal
        show={showCommentModal}
        onClose={() => setShowCommentModal(false)}
        comments={comments}
        newComment={newComment}
        setNewComment={setNewComment}
        handlePostComment={handlePostComment}
        editingCommentId={editingCommentId}
        editText={editText}
        setEditText={setEditText}
        handleSaveEdit={handleSaveEdit}
        handleCancelEdit={handleCancelEdit}
        handleEditClick={handleEditClick}
        handleDeleteComment={handleDeleteComment}
      />
    </>
  );
};

export default PostCard;
