"""
Deduplication package for tennis court slot scraping.

This package provides Redis-based deduplication services to prevent
reprocessing of recently seen court slots.
"""

from .redis_deduplicator import RedisDeduplicator

__all__ = ['RedisDeduplicator'] 