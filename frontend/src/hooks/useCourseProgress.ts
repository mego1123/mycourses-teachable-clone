// useCourseProgress — hook that sends progress updates to the backend every 15 seconds.
// Called by the video player on timeupdate events.
import { useEffect, useRef, useCallback } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { progressApi } from '../api/courseApi';

export function useCourseProgress(lessonId: string | undefined, courseId: string | undefined) {
  const queryClient = useQueryClient();
  const lastPositionRef = useRef(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const mutation = useMutation({
    mutationFn: (data: { videoPositionSec: number; completed: boolean }) =>
      progressApi.update(lessonId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['player-progress', courseId] });
    },
  });

  const reportProgress = useCallback((positionSec: number, durationSec: number, completed: boolean) => {
    if (!lessonId) return;
    lastPositionRef.current = positionSec;

    // Auto-mark as completed if watched 90% of the video
    const autoCompleted = !completed && durationSec > 0 && positionSec / durationSec > 0.9;

    mutation.mutate({
      videoPositionSec: Math.floor(positionSec),
      completed: completed || autoCompleted,
    });
  }, [lessonId, mutation]);

  // Start a heartbeat that reports progress every 15 seconds
  const startHeartbeat = useCallback((getCurrentPosition: () => number, getDuration: () => number) => {
    if (intervalRef.current) clearInterval(intervalRef.current);

    intervalRef.current = setInterval(() => {
      const position = getCurrentPosition();
      const duration = getDuration();
      if (position > 0) {
        reportProgress(position, duration, false);
      }
    }, 15000); // 15 seconds
  }, [reportProgress]);

  // Stop the heartbeat
  const stopHeartbeat = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => stopHeartbeat();
  }, [stopHeartbeat]);

  return { reportProgress, startHeartbeat, stopHeartbeat };
}
