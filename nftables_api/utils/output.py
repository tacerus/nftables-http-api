import json

def output_post(status, message=None):
  output = {
    'status': status
  }

  if message:
    output.update(
      {
        'message': message
      }
    )

  return json.dumps(output)
