import os
from dotenv import load_dotenv

# Load environment variables from .env file if it exists (for local development)
# This should be called early, especially if other variables might depend on it.
load_dotenv()

DIALOGFLOW_PROJECT_ID = os.getenv("DIALOGFLOW_PROJECT_ID")
DIALOGFLOW_AGENT_ID = os.getenv("DIALOGFLOW_AGENT_ID")
# Default to 'global' for location if not explicitly set, as it's a common default.
DIALOGFLOW_LOCATION_ID = os.getenv("DIALOGFLOW_LOCATION_ID", "global")
# Default to 'en-US' for language code if not explicitly set.
DIALOGFLOW_LANGUAGE_CODE = os.getenv("DIALOGFLOW_LANGUAGE_CODE", "en-US")

# GOOGLE_APPLICATION_CREDENTIALS should be set in the server's environment directly
# for the Google Cloud client libraries to automatically find and use service account keys.
# This script checks if it's set and provides a warning to the developer if not.
GOOGLE_APP_CREDS = os.getenv("GOOGLE_APPLICATION_CREDENTIALS")

# Print warnings if essential Dialogflow configuration is missing.
# These warnings are primarily for developers during setup and local execution.
# In a production deployment, these should be configured correctly via the environment.
if not DIALOGFLOW_PROJECT_ID:
    print("Warning: DIALOGFLOW_PROJECT_ID environment variable not set. NLUService may not function correctly with Dialogflow CX.")
if not DIALOGFLOW_AGENT_ID:
    print("Warning: DIALOGFLOW_AGENT_ID environment variable not set. NLUService may not function correctly with Dialogflow CX.")
if not GOOGLE_APP_CREDS:
    print("Warning: GOOGLE_APPLICATION_CREDENTIALS environment variable not set. NLUService will likely fail to authenticate with Google Cloud Dialogflow CX.")
else:
    # Optionally print the path to the credentials file if set, for verification (be careful with logging sensitive paths).
    # print(f"Info: GOOGLE_APPLICATION_CREDENTIALS is set to: {GOOGLE_APP_CREDS}")
    pass


# Example of other NLU related configurations that could be added:
# NLU_PROVIDER = os.getenv("NLU_PROVIDER", "dialogflow_cx") # To switch between NLU providers
# DEFAULT_CONFIDENCE_THRESHOLD = float(os.getenv("DEFAULT_CONFIDENCE_THRESHOLD", "0.3"))
# CACHE_NLU_RESULTS = os.getenv("CACHE_NLU_RESULTS", "false").lower() == "true"
