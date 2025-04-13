import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { jwtDecode } from "jwt-decode";
import PostCard from "../components/PostCard";
import "bootstrap/dist/css/bootstrap.min.css";

const Profile = () => {
  const [user, setUser] = useState(null);
  const [userPosts, setUserPosts] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchUserData = async () => {
      try {
        const token = localStorage.getItem("jwtToken");
        const decoded = jwtDecode(token);
        const loggedInUserId = decoded.sub;

        const [userRes, postsRes] = await Promise.all([
          axios.get(`http://localhost:8080/v1/users/${loggedInUserId}`, {
            headers: { Authorization: `Bearer ${token}` },
          }),
          axios.get(
            `http://localhost:8080/v1/users/${loggedInUserId}/allposts`,
            {
              headers: { Authorization: `Bearer ${token}` },
            }
          ),
        ]);

        setUser(userRes.data.data);
        setUserPosts(postsRes.data.data || []);
      } catch (error) {
        console.error("Error fetching profile data:", error);
      }
    };

    fetchUserData();
  }, []);

  if (!user) {
    return <h3 className="text-center">Loading...</h3>;
  }

  return (
    <div className="container mt-4">
      {/* Top Half: Profile Info */}
      <h2 className="text-center mb-4">Profile</h2>
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
              <strong>Followers</strong>
              <p>{user.follower_count || 0}</p>
            </div>
            <div>
              <strong>Following</strong>
              <p>{user.following_count || 0}</p>
            </div>
          </div>
          <button
            className="btn btn-primary w-100 mt-3"
            onClick={() => navigate("/saved-posts")}
          >
            Saved Posts
          </button>
        </div>
      </div>

      {/* Bottom Half: User's Own Posts */}
      <h4 className="text-center mb-3">Your Posts</h4>
      {userPosts.length === 0 ? (
        <p className="text-center text-muted">You haven't posted anything yet.</p>
      ) : (
        <div className="row">
          {userPosts.map((post) => (
            <div className="col-md-4" key={post.id}>
              <PostCard post={post} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Profile;
