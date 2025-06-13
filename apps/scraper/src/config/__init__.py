"""
Configuration management for the Tennis Booker scraper service.

This module provides configuration loading from JSON files and environment variables,
following the unified configuration structure defined in the project.
"""

from .config import Config, load_config, get_config

__all__ = ['Config', 'load_config', 'get_config'] 