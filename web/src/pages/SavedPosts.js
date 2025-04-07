import React, { useState, useEffect } from "react";
import axios from "axios";
import PostCard from "../components/PostCard";
import "bootstrap/dist/css/bootstrap.min.css";

const SavedPosts = () => {
  const [savedPosts, setSavedPosts] = useState([]);

  useEffect(() => {
    const fetchSavedPosts = async () => {
      try {
        const token = localStorage.getItem("jwtToken");
        const response = await axios.get("http://localhost:8080/v1/users/savedpost", {
          headers: { Authorization: `Bearer ${token}` },
        });

        setSavedPosts(response.data.data || []);
      } catch (error) {
        console.error("Error fetching saved posts:", error);
      }
    };

    fetchSavedPosts();
  }, []);

  return (
    <div className="row">
      {savedPosts.map((post) => (
        <div className="col-md-4" key={post.id}>
          <PostCard post={post} />
        </div>
      ))}
    </div>
  );
};

export default SavedPosts;
