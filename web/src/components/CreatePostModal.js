import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

const CreatePostModal = ({ show, onHide, initialData = {}, onSubmit }) => {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [tags, setTags] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    if (initialData) {
      setTitle(initialData.title || '');
      setContent(initialData.content || '');
      setTags((initialData.tags || []).join(', '));
    }
  }, [initialData]);

  const handleSubmit = async (e) => {
    console.log('here')
    e.preventDefault();
    const payload = {
      title,
      content,
      tags: tags.split(',').map(tag => tag.trim()).filter(tag => tag)
    };

    const token = localStorage.getItem('jwtToken');

    try {
      if (initialData) {
        // Edit mode
        await axios.patch(`http://localhost:8080/v1/posts/${initialData.id}`, payload, {
          headers: { Authorization: `Bearer ${token}` }
        });
      } else {
        // Create mode
        await axios.post('http://localhost:8080/v1/posts/', payload, {
          headers: { Authorization: `Bearer ${token}` }
        });
      }

      if (onSubmit) onSubmit();
      onHide();
      navigate('/home');
    } catch (error) {
      console.error('Error submitting post:', error);
    }
  };

  return (
    <div className={`modal fade ${show ? 'show d-block' : ''}`} tabIndex="-1" role="dialog" style={{ backgroundColor: 'rgba(0,0,0,0.5)' }}>
      <div className="modal-dialog modal-dialog-centered" role="document">
        <div className="modal-content">
          <div className="modal-header">
            <h5 className="modal-title">{initialData ? 'Edit Post' : 'Create New Post'}</h5>
            <button type="button" className="btn-close" onClick={onHide} aria-label="Close"></button>
          </div>
          <form onSubmit={handleSubmit}>
            <div className="modal-body">
              <div className="mb-3">
                <label htmlFor="postTitle" className="form-label">Title</label>
                <input
                  type="text"
                  className="form-control"
                  id="postTitle"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  required
                />
              </div>
              <div className="mb-3">
                <label htmlFor="postContent" className="form-label">Content</label>
                <textarea
                  className="form-control"
                  id="postContent"
                  rows="4"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  required
                ></textarea>
              </div>
              <div className="mb-3">
                <label htmlFor="postTags" className="form-label">Tags (comma separated)</label>
                <input
                  type="text"
                  className="form-control"
                  id="postTags"
                  value={tags}
                  onChange={(e) => setTags(e.target.value)}
                />
              </div>
            </div>
            <div className="modal-footer">
              <button type="button" className="btn btn-secondary" onClick={onHide}>Cancel</button>
              <button type="submit" className="btn btn-primary">
                {initialData ? 'Update Post' : 'Add Post'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default CreatePostModal;
