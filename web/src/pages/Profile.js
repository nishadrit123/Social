import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import {jwtDecode} from "jwt-decode";
import "bootstrap/dist/css/bootstrap.min.css";

const Profile = () => {
  const [user, setUser] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchUserProfile = async () => {
      try {
        const token = localStorage.getItem("jwtToken");
        const decoded = jwtDecode(token);
        const loggedInUserId = decoded.sub;

        const response = await axios.get(
          `http://localhost:8080/v1/users/${loggedInUserId}`,
          { headers: { Authorization: `Bearer ${token}` } }
        );
        setUser(response.data.data);
      } catch (error) {
        console.error("Error fetching user profile:", error);
      }
    };

    fetchUserProfile();
  }, []);

  if (!user) {
    return <h3 className="text-center">Loading...</h3>;
  }

  return (
    <div className="container mt-4">
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
          <button className="btn btn-primary w-100 mt-3" onClick={() => navigate("/saved-posts")}>Saved Posts</button>
        </div>
      </div>
    </div>
  );
};

export default Profile;
