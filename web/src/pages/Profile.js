import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import axios from "axios";
import { jwtDecode } from "jwt-decode";
import PostCard from "../components/PostCard";
import CreatePostModal from "../components/CreatePostModal";
import LikedUsersModal from "../components/LikedUsersModal";
import "bootstrap/dist/css/bootstrap.min.css";
import { FiPlus } from "react-icons/fi";

const Profile = () => {
  const { userId } = useParams();
  const [user, setUser] = useState(null);
  const [userPosts, setUserPosts] = useState([]);
  const [followers, setFollowers] = useState([]);
  const [followings, setFollowings] = useState([]);
  const [showFollowersModal, setShowFollowersModal] = useState(false);
  const [showFollowingsModal, setShowFollowingsModal] = useState(false);
  const [isFollowing, setIsFollowing] = useState(false);
  const [followLoading, setFollowLoading] = useState(false);
  const [showCreatePostModal, setShowCreatePostModal] = useState(false);
  const [editingPost, setEditingPost] = useState(null);

  const token = localStorage.getItem("jwtToken");
  const decoded = jwtDecode(token);
  const loggedInId = decoded.sub;
  const isOwnProfile = !userId || userId === String(loggedInId);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchUserData = async () => {
      try {
        const token = localStorage.getItem("jwtToken");
        const decoded = jwtDecode(token);
        const loggedInId = decoded.sub;
        const profileId = userId || loggedInId;

        const [userRes, postsRes] = await Promise.all([
          axios.get(`http://localhost:8080/v1/users/${profileId}`, {
            headers: { Authorization: `Bearer ${token}` },
          }),
          axios.get(`http://localhost:8080/v1/users/${profileId}/allposts`, {
            headers: { Authorization: `Bearer ${token}` },
          }),
        ]);

        setUser(userRes.data.data);
        setUserPosts(postsRes.data.data || []);
        setIsFollowing(userRes.data.data.is_already_following || false);
      } catch (error) {
        console.error("Error fetching profile data:", error);
      }
    };

    fetchUserData();
  }, [userId]);

  const handleFollowersClick = async () => {
    try {
      const token = localStorage.getItem("jwtToken");
      const profileId = userId || loggedInId;
      const res = await axios.get(
        `http://localhost:8080/v1/users/${profileId}/allfollowers`,
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setFollowers(res.data.data || []);
      setShowFollowersModal(true);
    } catch (error) {
      console.error("Failed to fetch followers", error);
    }
  };

  const handleFollowingsClick = async () => {
    try {
      const token = localStorage.getItem("jwtToken");
      const profileId = userId || loggedInId;
      const res = await axios.get(
        `http://localhost:8080/v1/users/${profileId}/allfollowings`,
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setFollowings(res.data.data || []);
      setShowFollowingsModal(true);
    } catch (error) {
      console.error("Failed to fetch followings", error);
    }
  };

  const handleFollowToggle = async () => {
    try {
      setFollowLoading(true);
      const endpoint = isFollowing
        ? `http://localhost:8080/v1/users/${userId}/unfollow`
        : `http://localhost:8080/v1/users/${userId}/follow`;

      await axios.put(
        endpoint,
        {},
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );

      // Optimistically update follow status and follower count
      setIsFollowing((prev) => !prev);
      setUser((prevUser) => ({
        ...prevUser,
        follower_count:
          Number(prevUser.follower_count) + (isFollowing ? -1 : 1),
      }));
    } catch (error) {
      console.error("Error toggling follow state:", error);
    } finally {
      setFollowLoading(false);
    }
  };

  const handleDeletePost = async (postId) => {
    try {
      const token = localStorage.getItem("jwtToken");
      await axios.delete(`http://localhost:8080/v1/posts/${postId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      // Update the local state to remove the deleted post
      setUserPosts((prevPosts) =>
        prevPosts.filter((post) => post.id !== postId)
      );
      navigate('/home');
    } catch (error) {
      console.error("Failed to delete post:", error);
    }
  };

  const handleEditPost2 = (post) => {
    setEditingPost(post);
    setShowCreatePostModal(true);
  };

  const handlePostSubmit = async (postData) => {
    try {
      const token = localStorage.getItem("jwtToken");
      if (editingPost) {
        // Editing existing post
        await axios.patch(
          `http://localhost:8080/v1/posts/${editingPost.id}`,
          postData,
          {
            headers: { Authorization: `Bearer ${token}` },
          }
        );
        setUserPosts((prevPosts) =>
          prevPosts.map((post) =>
            post.id === editingPost.id ? { ...post, ...postData } : post
          )
        );
      } else {
        // Creating new post
        const response = await axios.post(
          "http://localhost:8080/v1/posts/",
          postData,
          {
            headers: { Authorization: `Bearer ${token}` },
          }
        );
        setUserPosts((prevPosts) => [response.data.data, ...prevPosts]);
      }
    } catch (error) {
      console.error("Failed to submit post:", error);
    } finally {
      setEditingPost(null);
      setShowCreatePostModal(false);
    }
  };

  if (!user) {
    return <h3 className="text-center">Loading...</h3>;
  }

  return (
    <div className="container mt-4">
      <h2 className="text-center mb-4">Profile</h2>
      {isOwnProfile && (
        <button
          className="btn btn-outline-primary d-flex align-items-center gap-2 px-3 py-2 rounded"
          onClick={() => {
            setEditingPost(null);
            setShowCreatePostModal(true);
          }}
          aria-label="Create Post"
          title="Create Post"
        >
          <FiPlus size={18} />
        </button>
      )}
      <div className="card mx-auto mb-4" style={{ maxWidth: "600px" }}>
        <div className="card-body">
          <h5 className="card-title text-center">{user.username}</h5>
          <p className="text-muted text-center">{user.email}</p>
          <div className="d-flex justify-content-around mt-3">
            <div>
              <strong>Posts</strong>
              <p>{user.post_count || 0}</p>
            </div>
            <div>
              <strong
                style={{ cursor: "pointer" }}
                onClick={handleFollowersClick}
              >
                Followers
              </strong>
              <p>{user.follower_count || 0}</p>
            </div>
            <div>
              <strong
                style={{ cursor: "pointer" }}
                onClick={handleFollowingsClick}
              >
                Following
              </strong>
              <p>{user.following_count || 0}</p>
            </div>
          </div>
  
          {isOwnProfile ? (
            <button
              className="btn btn-primary w-100 mt-3"
              onClick={() => navigate("/saved-posts")}
            >
              Saved Posts
            </button>
          ) : (
            <div className="d-flex gap-2 mt-3">
              <button className="btn btn-outline-secondary w-50">TBH</button>
              <button
                className={`btn w-50 ${
                  isFollowing ? "btn-danger" : "btn-outline-primary"
                }`}
                onClick={handleFollowToggle}
                disabled={followLoading}
              >
                {followLoading
                  ? "Processing..."
                  : isFollowing
                  ? "Unfollow"
                  : "Follow"}
              </button>
            </div>
          )}
        </div>
      </div>
  
      <h4 className="text-center mb-3">
        {isOwnProfile ? "Your Posts" : `${user.username}'s Posts`}
      </h4>
      {userPosts.length === 0 ? (
        <p className="text-center text-muted">
          {isOwnProfile
            ? "You haven't posted anything yet."
            : "This user hasn't posted anything yet."}
        </p>
      ) : (
        <div className="row">
          {userPosts.map((post) => (
            <div className="col-md-4" key={post.id}>
              <PostCard
                post={post}
                isOwnProfile={isOwnProfile}
                onEdit={handleEditPost2}
                onDelete={handleDeletePost}
              />
            </div>
          ))}
        </div>
      )}
  
      <LikedUsersModal
        show={showFollowersModal}
        likedUsers={followers}
        onClose={() => setShowFollowersModal(false)}
        title={"Followers"}
        emptytitle={"followers"}
      />
      <LikedUsersModal
        show={showFollowingsModal}
        likedUsers={followings}
        onClose={() => setShowFollowingsModal(false)}
        title={"Following"}
        emptytitle={"followings"}
      />
      <CreatePostModal
        show={showCreatePostModal}
        onHide={() => {
          setShowCreatePostModal(false);
          setEditingPost(null);
        }}
        initialData={editingPost}
        onSubmit={handlePostSubmit}
      />
    </div>
  );
  
};

export default Profile;
