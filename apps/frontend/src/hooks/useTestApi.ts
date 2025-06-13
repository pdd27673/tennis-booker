import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

// Mock API functions
const fetchPosts = async () => {
  const response = await fetch('https://jsonplaceholder.typicode.com/posts?_limit=5')
  if (!response.ok) {
    throw new Error('Failed to fetch posts')
  }
  return response.json()
}

const fetchPost = async (id: number) => {
  const response = await fetch(`https://jsonplaceholder.typicode.com/posts/${id}`)
  if (!response.ok) {
    throw new Error('Failed to fetch post')
  }
  return response.json()
}

const createPost = async (post: { title: string; body: string; userId: number }) => {
  const response = await fetch('https://jsonplaceholder.typicode.com/posts', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(post),
  })
  if (!response.ok) {
    throw new Error('Failed to create post')
  }
  return response.json()
}

// React Query hooks
export const usePosts = () => {
  return useQuery({
    queryKey: ['posts'],
    queryFn: fetchPosts,
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}

export const usePost = (id: number) => {
  return useQuery({
    queryKey: ['post', id],
    queryFn: () => fetchPost(id),
    enabled: !!id, // Only run if id is provided
  })
}

export const useCreatePost = () => {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: createPost,
    onSuccess: () => {
      // Invalidate and refetch posts after creating a new one
      queryClient.invalidateQueries({ queryKey: ['posts'] })
    },
  })
}

// Types
export interface Post {
  id: number
  title: string
  body: string
  userId: number
} 